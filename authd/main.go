package main

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/globalsign/mgo"
	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logger        *zap.Logger
	version       = "unknown"
	dbAddr        = "mongodb://myuser:pass@47.100.31.117:27017/mydb"
	collection    *mgo.Collection
	signSecretKey = []byte("my-secret-20200202")
)

// User represents a user
type User struct {
	ID        string `json:"_id"`
	Name      string `json:"name"`
	HashPass  string `json:"hashpass"`
	Activated bool   `json:"activated"`
}

func createToken(user string) (string, error) {
	// Create a new token object, specifying signing method and the claims
	// you would like it to contain.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		Subject: user,
		// set the expire time
		// see http://tools.ietf.org/html/draft-ietf-oauth-json-web-token-20#section-4.1.4
		ExpiresAt: time.Now().Add(time.Hour * 12).Unix(),
	})

	// Sign and get the complete encoded token as a string using the secret
	return token.SignedString(signSecretKey)
}

func validatePassword(user, pass string) error {
	u := User{}

	if err := collection.FindId(user).One(&u); err != nil {
		return err
	}
	if u.HashPass != pass {
		return errors.New("invalid password")
	}

	return nil
}

func authHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	user := r.FormValue("user")
	pass := r.FormValue("pass")

	logger.Info("authenticate",
		zap.String("user", user),
		zap.String("pass", pass),
	)

	if err := validatePassword(user, pass); err != nil {
		w.WriteHeader(http.StatusPreconditionFailed)
		w.Write([]byte("invalid user/passwd"))
		logger.Info("validate password",
			zap.String("err", err.Error()),
		)
		return
	}

	tokenString, err := createToken(user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error while signing token"))
		logger.Info("signing token",
			zap.String("err", err.Error()),
		)
		return
	}

	w.Header().Set("Content-Type", "application/jwt")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, tokenString)
}

func pingHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Fprintf(w, "hello, %s!\n", ps.ByName("name"))
}

// init logger
func initLogger(debug bool) (logger *zap.Logger) {
	var err error
	var zcfg zap.Config
	if debug {
		zcfg = zap.NewDevelopmentConfig()
	} else {
		zcfg = zap.NewProductionConfig()
		// Change default(1578990857.105345) timeFormat to 2020-01-14T16:35:34.851+0800
		zcfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}
	if os.Getenv("LOGGER") == "file" {
		filename := filepath.Base(os.Args[0])
		zcfg.OutputPaths = []string{
			filepath.Join("/tmp", filename),
		}
	}

	logger, err = zcfg.Build()
	if err != nil {
		panic(fmt.Sprintf("initLooger error %s", err))
	}

	zap.ReplaceGlobals(logger)
	return
}

func initDB() *mgo.Collection {
	dI, err := mgo.ParseURL(dbAddr)
	if err != nil {
		panic(err)
	}

	session, err := mgo.DialWithInfo(dI)
	if err != nil {
		panic(err)
	}

	db := session.DB(dI.Database)
	return db.C("users")
}

func main() {
	port := flag.Uint("port", 80, "port")
	debug := flag.Bool("debug", false, "debug")
	ver := flag.Bool("version", false, "version")
	flag.Parse()

	if *ver {
		fmt.Println(version)
		return
	}

	logger = initLogger(*debug)
	defer logger.Sync()

	collection = initDB()

	collection.Find("")

	router := httprouter.New()
	router.POST("/auth", authHandler)
	router.GET("/ping/:id", pingHandler)

	addr := fmt.Sprintf(":%d", *port)

	logger.Info("starting",
		zap.String("addr", addr),
	)
	if err := http.ListenAndServe(addr, router); err != nil {
		logger.Fatal("ListenAndServe",
			zap.String("addr", addr),
			zap.String("err", err.Error()),
		)
	}
}
