package models

import (
	"encoding/json"
	"fmt"
	"github.com/ungerik/go-gravatar"
	"io/ioutil"
	"labix.org/v2/mgo/bson"
	"net/http"
	"strings"
	"time"
)

type FacebookProfile struct {
	Id          string           //1681448470
	First_name  string           //Radu
	Middle_name string           //Ioan
	Last_name   string           //Fericean
	Username    string           //radu.fericean
	Birthday    string           //07\/10\/1978
	Gender      string           //male
	Email       string           //fericean\u0040gmail.com
	Timezone    int              //2
	Locale      string           //en_US
	Location    FacebookLocation //{"id":"107982459236366","name":"Timisoara, Romania"},
}

func (fbp *FacebookProfile) GetBirthdate() (time.Time, error) {
	return time.Parse("01/02/2006", fbp.Birthday)
}

func (fbp *FacebookProfile) GetGender() string {
	if fbp.Gender == "male" {
		return "m"
	}
	return "f"
}

func (fbp *FacebookProfile) GetLocation() (string, string) {
	loc := strings.Split(fbp.Location.Name, ",")
	if len(loc) != 2 {
		return "", ""
	}
	return strings.TrimSpace(loc[0]), strings.TrimSpace(loc[1])
}

//Login validates and returns a user object if they exist in the database.
func LoginWithFacebook(ctx *Context, fbp *FacebookProfile) (u *User, redirect string, err error) {
	// check if facebook profile has no email set
	if fbp.Email == "" {
		fbp.Email = fbp.Username + "@facebook.com"
	}
	err = ctx.C("users").Find(bson.M{"email": fbp.Email}).One(&u)
	if err != nil {
		bday, err := fbp.GetBirthdate()
		if err != nil {
			bday = time.Time{}
		}
		city, country := fbp.GetLocation()
		u = &User{
			Id:        bson.NewObjectId(),
			Email:     fbp.Email,
			FirstName: fbp.First_name + " " + fbp.Middle_name,
			LastName:  fbp.Last_name,
			Country:   country,
			Location:  city,
			BirthDate: bday,
			Gender:    fbp.GetGender(),
			FbId:      fbp.Id,
		}
		// set avatar
		if resp, err := http.Get(fmt.Sprintf("https://graph.facebook.com/%s?fields=picture", u.FbId)); err == nil {
			defer resp.Body.Close()
			if profile, err := ioutil.ReadAll(resp.Body); err == nil {
				a := struct {
					Picture struct{ Data struct{ Url string } }
				}{}
				if err = json.Unmarshal(profile, &a); err == nil {
					Cache.Set(u.FbId, a.Picture.Data.Url, 0)
					u.Avatar = a.Picture.Data.Url
				}
			}
		}
		if u.Avatar == "" {
			u.Avatar = gravatar.UrlSize(u.Email, 80)
		}
		err = ctx.C("users").Insert(u)
		if err != nil {
			return nil, "login", err
		}
		redirect = "profile"
	}
	redirect = "index"
	return
}

type GoogleProfile struct {
	Id          string //105466292173351316309
	Name        string //Radu Ioan Fericean
	Given_name  string //Radu Ioan
	Family_name string //Fericean
	Link        string //https://plus.google.com/105466292173351316309
	Birthday    string //0000-07-10
	Gender      string //male
	Email       string //fericean@gmail.com
	Locale      string //ro
	Picture     string //https://lh5.googleusercontent.com/-Ie8_S-f7keQ/AAAAAAAAAAI/AAAAAAAAA0k/qTcX65urSV4/photo.jpg
}

func (gp *GoogleProfile) GetBirthdate() (time.Time, error) {
	return time.Parse("2006-01-02", gp.Birthday)
}

func (gp *GoogleProfile) GetGender() string {
	if gp.Gender == "male" {
		return "m"
	}
	return "f"
}

//Login validates and returns a user object if they exist in the database.
func LoginWithGoogle(ctx *Context, gp *GoogleProfile) (u *User, redirect string, err error) {
	err = ctx.C("users").Find(bson.M{"email": gp.Email}).One(&u)
	if err != nil {
		bday, err := gp.GetBirthdate()
		if err != nil {
			bday = time.Time{}
		}
		u = &User{
			Id:        bson.NewObjectId(),
			Email:     gp.Email,
			FirstName: gp.Given_name,
			LastName:  gp.Family_name,
			BirthDate: bday,
			Gender:    gp.GetGender(),
			GlId:      gp.Id,
		}

		// set avatar
		u.Avatar = gp.Picture
		if u.Avatar == "" {
			u.Avatar = gravatar.UrlSize(u.Email, 80)
		}
		err = ctx.C("users").Insert(u)
		if err != nil {
			return nil, "login", err
		}
		redirect = "profile"
	}
	redirect = "index"
	return
}
