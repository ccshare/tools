package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/gobike/envflag"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/crypto/ssh"
)

var (
	version          = "unknown"
	failureKeyPrefix = "side:failures:"
	logMarkerKey     = "side:logger:marker"
	logMarker        = ""
	tmpDir           = os.TempDir()
	logger           *zap.Logger
	storeDb          *redis.Client
	storeFd          *os.File
)

func main() {
	user := flag.String("user", "root", "ssh user name")
	passwd := flag.String("passwd", "ChangeMe", "ssh user passwd")
	server := flag.String("server", "192.168.55.2:22", "ssh server ip:port")
	store := flag.String("store", "fs", "where to store log(fs or redis://127.0.0.1:6379/0)")
	debug := flag.Bool("debug", false, "debug log level")
	ver := flag.Bool("version", false, "show version")
	cmd := flag.String("cmd", "tail -q -n +1 -F --max-unchanged-stats=5", "cmd to collect log")
	file := flag.String("file", "/var/log/vipr/emcvipr-object/dataheadsvc-access.log", "remote log file name")

	envflag.Parse()
	if *ver {
		fmt.Println(version)
		return
	}
	logger = initLogger(*debug)
	defer logger.Sync()

	client, err := newSSHClient(*user, *passwd, *server)
	if err != nil {
		logger.Fatal("new ssh client failed",
			zap.String("err", err.Error()),
		)
	}

	if *store == "fs" { // use fs as log store
		storeFilename := filepath.Join(tmpDir, fmt.Sprintf("%s-%s.log", filepath.Base(*file), *server))
		logMarker = findMarkerFromFile(storeFilename)
		storeFd, err = os.OpenFile(storeFilename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			logger.Fatal("open store file failed",
				zap.String("err", err.Error()),
			)
		}
		defer storeFd.Close()
		logger.Info("got marker from file",
			zap.String("marker", logMarker),
		)
	} else { // use redis as log store
		storeDb, logMarker, err = dbInit(*store, logMarkerKey, *server)
		if err != nil {
			logger.Fatal("init db failed",
				zap.String("err", err.Error()),
			)
		}
		defer storeDb.Close()
		logger.Info("got marker from db",
			zap.String("marker", logMarker),
		)
	}

	// collect archived logs 2 times to make sure new rotate log collected
	for i := 0; i < 2; i++ {
		logfiles, err := searchArchivedLogfile(client, *server, *file, logMarker)
		if err != nil {
			logger.Error("collect archived log failed",
				zap.Int("index", i),
				zap.String("err", err.Error()),
			)
		}
		// collect(zcat) archived logs
		for _, v := range logfiles {
			logMarker, err = collectLog(client, *server, "zcat", v, logMarker)
			if err != nil {
				logger.Error("collect log error",
					zap.Int("index", i),
					zap.String("logfile", v),
					zap.String("err", err.Error()),
				)
			} else {
				logger.Info("collect log success",
					zap.Int("index", i),
					zap.String("logfile", v),
					zap.String("marker", logMarker),
				)
			}
		}
	}

	// loop collect(tail) latest log
	marker := logMarker
	for {
		logger.Info("tail log",
			zap.String("file", *file),
			zap.String("marker", marker),
		)
		marker, err = collectLog(client, *server, *cmd, *file, marker)
		if err != nil {
			logger.Error("tail log error",
				zap.String("cmd", *cmd),
				zap.String("file", *file),
				zap.String("err", err.Error()),
			)
		}
	}
	// end of main
}

// initLogger init logger
func initLogger(debug bool) *zap.Logger {
	zcfg := zap.NewProductionConfig()
	// Change default(1578990857.105345) timeFormat to 2020-01-14T16:35:34.851+0800
	zcfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	if debug {
		zcfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}
	if os.Getenv("LOGGER") == "file" {
		filename := filepath.Base(os.Args[0])
		zcfg.OutputPaths = []string{
			filepath.Join(tmpDir, fmt.Sprintf("%s.log", filename)),
		}
	}
	logger, err := zcfg.Build()
	if err != nil {
		panic(fmt.Sprintf("initLooger error %s", err))
	}
	zap.ReplaceGlobals(logger)
	return logger
}

func newSSHClient(user, passwd, server string) (*ssh.Client, error) {
	config := ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(passwd),
			ssh.KeyboardInteractive(func(user, instruction string, questions []string, echos []bool) ([]string, error) {
				// Just send the password back for all questions
				answers := make([]string, len(questions))
				for i := range answers {
					answers[i] = passwd // replace this
				}
				return answers, nil
			}),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		//HostKeyCallback: ssh.FixedHostKey(hostKey),
	}

	return ssh.Dial("tcp", server, &config)
}

func searchArchivedLogfile(client *ssh.Client, serverAddr, filename, marker string) ([]string, error) {
	session, err := client.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	buff := &bytes.Buffer{}
	session.Stdout = buff
	cmd := fmt.Sprintf("ls -1 %s*gz", filename)
	err = session.Run(cmd)
	if err != nil {
		return nil, fmt.Errorf("run ls -l logfile*gz failed, %w", err)
	}

	allLogfiles := []string{}
	scanner := bufio.NewScanner(buff)
	for scanner.Scan() {
		allLogfiles = append(allLogfiles, scanner.Text())
	}
	sort.Strings(allLogfiles)
	if len(marker) < len("2006-01-02 15:04:05") {
		return allLogfiles, fmt.Errorf("invalid marker: %s", marker)
	}
	t, err := time.Parse("2006-01-02 15:04:05", marker[0:19])
	if err != nil {
		return allLogfiles, fmt.Errorf("invalid marker: %s, %w", marker, err)
	}
	marker = t.Format("20060102-150405")

	logfiles := []string{}
	for _, v := range allLogfiles {
		fields := strings.Split(v, ".")
		if len(fields) != 4 || !strings.Contains(fields[2], "-") {
			logger.Warn("invalid archived filename",
				zap.String("filename", v),
			)
			continue
		}
		if fields[2] < marker {
			logger.Info("ignore archived file",
				zap.String("filename", v),
			)
			continue
		}
		logger.Info("hit archived log",
			zap.String("filename marker", marker),
			zap.String("filename", v),
		)
		logfiles = append(logfiles, v)
	}

	return logfiles, nil
}

func parseLog(r io.Reader, server, marker string) (newMarker string, err2 error) {
	newMarker = marker
	reader := bufio.NewReader(r)
	for {
		line, prefix, err := reader.ReadLine()
		if err != nil {
			err2 = err
			break
		}
		if prefix {
			logger.Warn("readline not finish",
				zap.Bool("prefix", prefix),
			)
			continue
		}
		fields := strings.Split(string(line), " ")
		if len(fields) < 2 || fields[1] == "1.0" || fields[1] == "date" {
			// ignore log header
			// Version: 1.0
			// #Fields: date time x-request-id s-ip c-ip ...
			logger.Info("skip log",
				zap.String("log", fields[0]),
			)
			continue
		} else if len(fields) > 12 {
			newMarker = fmt.Sprintf("%s %s", fields[0], fields[1])
			if newMarker <= marker {
				logger.Debug("skip log",
					zap.String("marker", newMarker),
				)
				continue
			}
			if fields[10] != "-" && (fields[7] == "PUT" || fields[7] == "POST") {
				if storeDb != nil {
					logger.Debug("got log",
						zap.String("id", fields[2]),
						zap.String("ak", fields[5]),
						zap.String("method", fields[7]),
						zap.String("bucket", fields[9]),
						zap.String("object", fields[10]),
					)
					// side:failures:accessKey (SET)
					bucketsKey := fmt.Sprintf("%s%s", failureKeyPrefix, fields[5])
					if err := storeDb.SAdd(bucketsKey, fields[9]).Err(); err != nil {
						logger.Error("write bucket to db failed",
							zap.String("dbkey", bucketsKey),
							zap.String("bucket", fields[9]),
							zap.String("err", err.Error()),
						)
					}
					// side:failures:accessKey:bucket (LIST)
					objectsKey := fmt.Sprintf("%s:%s", bucketsKey, fields[9])
					if err := storeDb.RPush(objectsKey, fields[10]).Err(); err != nil {
						logger.Error("write object to db failed",
							zap.String("dbkey", objectsKey),
							zap.String("bucket", fields[10]),
							zap.String("err", err.Error()),
						)
					}
					if err := storeDb.HSet(logMarkerKey, server, newMarker).Err(); err != nil {
						logger.Warn("write marker to db failed",
							zap.String("err", err.Error()),
						)
					}
				}
			}
			if storeFd != nil {
				storeFd.Write(line)
				storeFd.Write([]byte("\n"))
			}
		} else {
			logger.Warn("invalid log",
				zap.String("log", fields[0]),
			)
		}
	}
	return
}

func collectLog(client *ssh.Client, serverAddr, cmd, filename, marker string) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	outReader, err := session.StdoutPipe()
	if err != nil {
		return "", err
	}

	newMarker := marker
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func(r io.Reader, wg *sync.WaitGroup) {
		newMarker, err = parseLog(r, serverAddr, marker)
		if err != nil {
			if err != io.EOF {
				logger.Warn("parseLog error",
					zap.String("err", err.Error()),
				)
			} else {
				err = nil
			}
		}
		wg.Done()
	}(outReader, &wg)

	errReader, err := session.StderrPipe()
	if err != nil {
		return newMarker, err
	}
	go func(r io.Reader) {
		reader := bufio.NewReader(r)
		for {
			line, _, err := reader.ReadLine()
			if err != nil {
				if err == io.EOF {
					break
				}
				logger.Error("stderr error",
					zap.String("err", err.Error()),
				)
				continue
			}
			logger.Warn("got msg from stderr",
				zap.String("msg", string(line)),
			)
		}
	}(errReader)

	runCmd := fmt.Sprintf("%s %s", cmd, filename)
	logger.Info("collect log",
		zap.String("cmd", runCmd),
		zap.String("marker", marker),
	)
	err = session.Run(runCmd)

	wg.Wait()
	return newMarker, err
}
