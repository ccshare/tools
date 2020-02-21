package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-redis/redis/v7"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/crypto/ssh"
)

var (
	version          = "unknown"
	failureKeyPrefix = "side:failures:"
	logMarkerKey     = "logger:marker"
	logMarker        = ""
	tmpDir           = os.TempDir()
	logger           *zap.Logger
	dbClient         *redis.Client
	storeFd          *os.File
)

func main() {
	user := flag.String("u", "root", "user name")
	passwd := flag.String("p", "dawter", "user passwd")
	server := flag.String("s", "192.168.55.2:22", "ssh server")
	dbaddr := flag.String("db", "", "where to store log(default fs, or redis://127.0.0.1:6379/0)")
	debug := flag.Bool("debug", false, "debug log level")
	ver := flag.Bool("version", false, "show version")
	cmd := flag.String("cmd", "tail -q -n +1 -F --max-unchanged-stats=5", "remote cmd to run")
	remoteLogfile := flag.String("rlf", "/var/log/vipr/emcvipr-object/dataheadsvc-access.log", "remote log file name")

	flag.Parse()
	if *ver {
		fmt.Println(version)
		fmt.Println(tmpDir)
		return
	}
	logger = initLogger(*debug)
	defer logger.Sync()

	envInit()
	client, err := newSSHClient(*user, *passwd, *server)
	if err != nil {
		logger.Fatal("new ssh client failed",
			zap.String("err", err.Error()),
		)
	}

	if *dbaddr != "" {
		dbClient, logMarker, err = dbInit(*dbaddr, logMarkerKey, *server)
		if err != nil {
			logger.Fatal("init db failed",
				zap.String("err", err.Error()),
			)
		}
		defer dbClient.Close()
	} else {
		storeFilename := filepath.Join(tmpDir, fmt.Sprintf("%s-%s.log", filepath.Base(*remoteLogfile), *server))
		logMarker = findMarkerFromFile(storeFilename)
		storeFd, err = os.OpenFile(storeFilename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			logger.Fatal("open store file failed",
				zap.String("err", err.Error()),
			)
		}
		defer storeFd.Close()
	}

	fmt.Println("marker: ", logMarker)

	if err := tailNewLog(*server, client, *cmd, *remoteLogfile); err != nil {
		fmt.Printf("error: %s\n", err)
		logger.Error("run",
			zap.String("err", err.Error()),
		)
	}

}

func envInit() {
	// manually set time zone
	if tz := os.Getenv("TZ"); tz != "" {
		var err error
		time.Local, err = time.LoadLocation(tz)
		if err != nil {
			logger.Warn("error loading zoneinfo",
				zap.String("TZ", tz),
				zap.String("error", err.Error()),
			)
		}
	}
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

func collectOldLog(client *ssh.Client, marker string) error {

	return nil
}

func tailNewLog(addr string, client *ssh.Client, cmd, filename string) error {
	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	outReader, err := session.StdoutPipe()
	if err != nil {
		return err
	}
	go func(r io.Reader) {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			text := scanner.Text()
			fields := strings.Split(text, " ")
			if len(fields) < 2 || fields[1] == "1.0" || fields[1] == "date" {
				// ignore log header
				// Version: 1.0
				// #Fields: date time x-request-id s-ip c-ip
				continue
			} else if len(fields) > 12 {
				if fields[10] != "-" && (fields[7] == "PUT" || fields[7] == "POST") {
					if dbClient != nil {
						logger.Debug("got log",
							zap.String("id", fields[2]),
							zap.String("ak", fields[5]),
							zap.String("method", fields[7]),
							zap.String("bucket", fields[9]),
							zap.String("object", fields[10]),
						)
						// side:failures:accessKey (SET)
						bucketsKey := fmt.Sprintf("%s%s", failureKeyPrefix, fields[5])
						if err := dbClient.SAdd(bucketsKey, fields[9]).Err(); err != nil {
							logger.Error("write bucket to db failed",
								zap.String("dbkey", bucketsKey),
								zap.String("bucket", fields[9]),
								zap.String("err", err.Error()),
							)
						}
						// side:failures:accessKey:bucket (LIST)
						objectsKey := fmt.Sprintf("%s:%s", bucketsKey, fields[9])
						if err := dbClient.RPush(objectsKey, fields[10]).Err(); err != nil {
							logger.Error("write object to db failed",
								zap.String("dbkey", objectsKey),
								zap.String("bucket", fields[10]),
								zap.String("err", err.Error()),
							)
						}
						if err := dbClient.HSet(logMarkerKey, addr, fmt.Sprintf("%s %s", fields[0], fields[1])).Err(); err != nil {
							logger.Warn("write marker to db failed",
								zap.String("err", err.Error()),
							)
						}
					}
				}
				if storeFd != nil {
					storeFd.Write([]byte(text))
					storeFd.Write([]byte("\n"))
				}
			} else {
				logger.Warn("invalid log",
					zap.String("log", text),
				)
			}
		}
		if err := scanner.Err(); err != nil {
			logger.Error("stdout error",
				zap.String("err", err.Error()),
			)
		}
	}(outReader)

	errReader, err := session.StderrPipe()
	if err != nil {
		return err
	}
	go func(r io.Reader) {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			logger.Warn("got msg from stderr",
				zap.String("msg", scanner.Text()),
			)
		}
		if err := scanner.Err(); err != nil {
			logger.Error("stderr error",
				zap.String("err", err.Error()),
			)
		}
	}(errReader)

	return session.Run(fmt.Sprintf("%s %s", cmd, filename))
}
