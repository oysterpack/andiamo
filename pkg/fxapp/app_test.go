/*
 * Copyright (c) 2019 OysterPack, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package fxapp_test

import (
	"errors"
	"github.com/Masterminds/semver"
	"github.com/oklog/ulid"
	"github.com/oysterpack/partire-k8s/pkg/fxapp"
	"github.com/oysterpack/partire-k8s/pkg/ulidgen"
	"go.uber.org/fx"
	"log"
	"testing"
	"time"
)

type Bar string
type Baz string

func ProvideBar() Bar {
	return "bar"
}

func ProvideBaz() Baz {
	return "baz"
}

func InvokePrintBaz(baz Baz) {
	log.Printf("baz = %v", baz)
}

func InvokePrintBar(bar Bar) {
	log.Printf("bar = %v", bar)
}

type FooID ulid.ULID

func NewFooID() FooID {
	return FooID(ulidgen.MustNew())
}

func (id FooID) ProvideFooID() FooID {
	return id
}

func (id FooID) InvokeLogSelf() {
	log.Printf("InvokeLogSelf: %s", ulid.ULID(id))
}

func (_ FooID) InvokeLogInstanceID(id FooID) {
	log.Printf("InvokeLogInstanceID: %s", ulid.ULID(id))
}

type Subject struct{}

type Login func(credentials interface{}) (*Subject, error)

type PasswordLogin struct {
	fx.Out

	Login `name:"PasswordLogin"`
}

func ProvidePasswordLogin() PasswordLogin {
	return PasswordLogin{
		Login: func(credentials interface{}) (subject *Subject, e error) {
			return nil, errors.New("password login not implemented")
		},
	}
}

type MFALogin struct {
	fx.Out

	Login `name:"MFALogin"`
}

func ProvideMFALogin() MFALogin {
	return MFALogin{
		Login: func(credentials interface{}) (subject *Subject, e error) {
			return nil, errors.New("MFA login not implemented")
		},
	}
}

type MFALoginParam struct {
	fx.In

	Login `name:"MFALogin"`
}

type PasswordLoginParam struct {
	fx.In

	Login `name:"PasswordLogin"`
}

// demonstrates use of named dependencies
func InvokeLogin(login MFALoginParam) {
	subject, err := login.Login("credentials")
	if err != nil {
		log.Printf("login failed: %v\n", err)
		return
	}
	log.Printf("logged in: %v\n", subject)
}

type LoginCommands struct {
	fx.Out

	Login `group:"Login"`
}

// demonstrates use of named dependencies
func GroupMFALogin(login MFALoginParam) LoginCommands {
	return LoginCommands{
		Login: login.Login,
	}
}

// demonstrates use of named dependencies
func GroupPasswordLogin(login PasswordLoginParam) LoginCommands {
	return LoginCommands{
		Login: login.Login,
	}
}

type Logins struct {
	fx.In

	Logins []Login `group:"Login"`
}

// demonstrates use of named groups and how to grouped named dependencies
func GatherLogins(logins Logins) {
	log.Println("Login count = ", len(logins.Logins))
	for _, login := range logins.Logins {
		log.Println(login("credentials"))
	}
}

func TestAppBuilder(t *testing.T) {
	// Given an App descriptor
	desc, e := fxapp.NewDescBuilder().
		SetID(ulidgen.MustNew()).
		SetName("foo").
		SetVersion(semver.MustParse("0.1.0")).
		SetReleaseID(ulidgen.MustNew()).
		Build()

	if e != nil {
		t.Fatalf("*** app failed to build desc: %v", e)
	}
	t.Logf("%v", desc)

	fooID := NewFooID()
	app, e := fxapp.NewAppBuilder(desc).
		SetStartTimeout(30*time.Second).
		SetStopTimeout(60*time.Second).
		Constructors(
			ProvideBar,
			ProvideBaz,
			fooID.ProvideFooID,
			ProvidePasswordLogin,
			ProvideMFALogin,
			GroupMFALogin,
			GroupPasswordLogin,
		).
		Funcs(
			InvokePrintBaz,
			InvokePrintBar,
			fooID.InvokeLogInstanceID,
			fooID.InvokeLogSelf,
			InvokeLogin,
			GatherLogins,
		).
		Build()

	if e != nil {
		t.Fatalf("*** app failed to build app: %v", e)
	}
	t.Logf("%v", app)

	if app.StartTimeout() != 30*time.Second {
		t.Errorf("*** start timeout did not match: %v", app.StartTimeout())
	}

	if app.StopTimeout() != 60*time.Second {
		t.Errorf("*** stop timeout did not match: %v", app.StopTimeout())
	}
	if app.Err() != nil {
		t.Fatalf("*** since the app was successfully built, then it should have no error")
	}

}
