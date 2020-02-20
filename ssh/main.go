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
	logMarker        map[string]string
	logger           *zap.Logger
	dbClient         *redis.Client
)

func main() {
	user := flag.String("u", "root", "user name")
	passwd := flag.String("p", "dawter", "user passwd")
	server := flag.String("s", "192.168.55.2:22", "ssh server")
	dbaddr := flag.String("db", "redis://127.0.0.1:6379/0", "redis address")
	debug := flag.Bool("debug", false, "debug log level")
	ver := flag.Bool("version", false, "show version")
	cmd1 := flag.String("cmd1", "tail -q -n +1 -F --max-unchanged-stats=5 /var/log/vipr/emcvipr-object/dataheadsvc-access.log", "cmd 1 to tun")

	flag.Parse()
	if *ver {
		fmt.Println(version)
		return
	}
	logger = initLogger(*debug)
	defer logger.Sync()

	envInit()
	var err error
	dbClient, err = dbInit(*dbaddr)
	if err != nil {
		logger.Fatal("init db failed",
			zap.String("err", err.Error()),
		)
	}
	logMarker, err = dbClient.HGetAll(logMarkerKey).Result()
	if err != nil {
		logger.Fatal("read log marker failed",
			zap.String("err", err.Error()),
		)
	}

	if err := run(*user, *passwd, *server, *cmd1); err != nil {
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
			filepath.Join("/tmp", fmt.Sprintf("%s.log", filename)),
		}
	}

	logger, err := zcfg.Build()
	if err != nil {
		panic(fmt.Sprintf("initLooger error %s", err))
	}

	zap.ReplaceGlobals(logger)
	return logger
}

func run(user, passwd, server, cmd1 string) error {
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

	client, err := ssh.Dial("tcp", server, &config)
	if err != nil {
		return err
	}
	defer client.Close()

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
			if len(fields) > 12 && fields[1] != "date" {
				if fields[10] != "-" && (fields[7] == "PUT" || fields[7] == "POST") {
					logger.Info("got log",
						zap.String("id", fields[2]),
						zap.String("ak", fields[5]),
						zap.String("method", fields[7]),
						zap.String("bucket", fields[9]),
						zap.String("object", fields[10]),
					)
				}
			} else {
				logger.Warn("invalid log",
					zap.Strings("log", fields),
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

	return session.Run(cmd1)
}
