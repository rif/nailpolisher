package models

import (
	"encoding/gob"
	"encoding/json"
	"github.com/gorilla/pat"
	"github.com/gorilla/sessions"
	"github.com/kisielk/raven-go/raven"
	"github.com/pmylund/go-cache"
	"io/ioutil"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"launchpad.net/goamz/aws"
	"launchpad.net/goamz/s3"
	"log"
	"net/http"
	"os"
	"path"
	"time"
)

var (
	store                sessions.Store
	db_session           *mgo.Session
	database             string
	Router               *pat.Router
	COOKIE_NAME          = "cmo"
	BASE_DIR             = os.Getenv("OPENSHIFT_REPO_DIR")
	DATA_DIR             = os.Getenv("OPENSHIFT_DATA_DIR")
	MONGO_DB_NAME        = "cmo"
	S3_BASE              = "http://s3.amazonaws.com/"
	s3ProductionBucketId = "photos.lov3ly.me"
	s3TestBucketId       = "test.lov3ly.me"
	S3_URL               = ""
	s3BucketId           = ""
	s3AccessKey          = "xxx"
	s3SecretKey          = "xxx"
	S3Bucket             *s3.Bucket
	Cache                = cache.New(5*time.Minute, 30*time.Second)
	sentry               *raven.Client
	SENTRY_DSN           = "https://xxx:xxx@sentry-rif.rhcloud.com/2"
	Translations         map[string]Trans
)

func init() {
	var err error
	mongo_server := os.Getenv("OPENSHIFT_MONGODB_DB_URL")
	if mongo_server == "" {
		db_session, err = mgo.Dial("localhost")
		s3BucketId = s3TestBucketId
		S3_URL = S3_BASE + s3BucketId
	} else {
		db_session, err = mgo.Dial(mongo_server + MONGO_DB_NAME)
		s3BucketId = s3ProductionBucketId
		S3_URL = S3_BASE + s3BucketId
	}
	s3Connection := s3.New(aws.Auth{s3AccessKey, s3SecretKey}, aws.USEast)
	S3Bucket = s3Connection.Bucket(s3BucketId)
	if err != nil {
		log.Print("db connection: ", err)
	}
	// register bson's ObjectId with the gob for cookie encoding
	gob.Register(bson.ObjectId(""))
	gob.RegisterName("app/models.Flash", &Flash{"", ""})

	database = db_session.DB("").Name
	Router = pat.New()
	//create an index for the email field on the users collection
	if err := db_session.DB(database).C("users").EnsureIndex(mgo.Index{
		Key:    []string{"email"},
		Unique: true,
	}); err != nil {
		log.Print("context: ", err)
	}
	if err := db_session.DB(database).C("users").EnsureIndexKey("location"); err != nil {
		log.Print("context: ", err)
	}
	if err := db_session.DB(database).C("users").EnsureIndexKey("country"); err != nil {
		log.Print("context: ", err)
	}
	store = sessions.NewCookieStore([]byte("508a664e65427d3f91000001"))
	if sentry, err = raven.NewClient(SENTRY_DSN); err != nil {
		log.Print("could not connect to sentry: ", err)
	}
	loadTranslations()
}

type Context struct {
	Database *mgo.Database
	Session  *sessions.Session
	User     *User
	Data     map[string]interface{}
}

func (c *Context) Close() {
	c.Database.Session.Close()
}

//C is a convenience function to return a collection from the context database.
func (c *Context) C(name string) *mgo.Collection {
	return c.Database.C(name)
}

func NewContext(req *http.Request) (*Context, error) {
	sess, err := store.Get(req, COOKIE_NAME)
	ctx := &Context{
		Database: db_session.Clone().DB(database),
		Session:  sess,
		Data:     make(map[string]interface{}),
	}
	if err != nil { // if the above is still an error
		return ctx, err
	}

	//try to fill in the user from the session
	if uid, ok := sess.Values["user"].(bson.ObjectId); ok {
		e := ctx.C("users").Find(bson.M{"_id": uid}).One(&ctx.User)
		if ctx.User != nil {
			ctx.User.Password = []byte{}
			ctx.User.BirthDate = ctx.User.BirthDate.UTC()
		}
		if e != nil {
			Log("error finding user for cookie uid: ", err.Error())
		}
	}
	if _, ok := sess.Values["csrf_token"].(string); !ok {
		ctx.Session.Values["csrf_token"] = bson.NewObjectId().Hex()
	}
	return ctx, err
}

type Trans map[string]string

func loadTranslations() {
	langPath := path.Join(BASE_DIR, "src/app/langs")
	files, err := ioutil.ReadDir(langPath)
	if err != nil {
		Log("error reading langs dir: ", err.Error())
		return
	}
	Translations = make(map[string]Trans)
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		file, err := os.Open(path.Join(langPath, f.Name()))
		if err != nil {
			Log("error opening lang file: ", err.Error())
			break
		}
		data, err := ioutil.ReadAll(file)
		if err != nil {
			Log("error reading lang file: ", err.Error())
			break
		}
		trans := Trans{}
		json.Unmarshal(data, &trans)
		Translations[f.Name()] = trans
	}
}
