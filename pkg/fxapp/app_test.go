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
	"context"
	"errors"
	"fmt"
	"github.com/Masterminds/semver"
	"github.com/oklog/ulid"
	"github.com/oysterpack/partire-k8s/pkg/fxapp"
	"github.com/oysterpack/partire-k8s/pkg/ulidgen"
	"go.uber.org/fx"
	"log"
	"reflect"
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

func newDesc(name, version string) fxapp.Desc {
	desc, e := fxapp.NewDescBuilder().
		SetID(ulidgen.MustNew()).
		SetName(name).
		SetVersion(semver.MustParse(version)).
		SetReleaseID(ulidgen.MustNew()).
		Build()
	if e != nil {
		panic(e)
	}
	return desc
}

// - constructors can be registered with the app
// - functions can be registered with the app
//   - at least 1 function must be registered
// - app start and stop time outs can be configured
// - a new app instance is assigned a unique instance ID
func TestAppBuilder(t *testing.T) {
	// Given an App descriptor
	desc := newDesc("foo", "0.1.0")

	timeBeforeBuildingApp := time.Now()
	fooID := NewFooID()
	app, e := fxapp.NewAppBuilder(desc).
		SetStartTimeout(30*time.Second).
		SetStopTimeout(60*time.Second).
		Provide(
			ProvideBar,
			ProvideBaz,
			fooID.ProvideFooID,
		).
		Invoke(
			InvokePrintBaz,
			InvokePrintBar,
			fooID.InvokeLogInstanceID,
			fooID.InvokeLogSelf,
		).
		Provide(
			ProvidePasswordLogin,
			ProvideMFALogin,
			GroupMFALogin,
			GroupPasswordLogin,
		).
		Invoke(
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

	appInstanceID := app.InstanceID()
	// subtract 1 millisecond because the ULID time is only millisecond precision
	if ulid.Time(appInstanceID.ULID().Time()).Before(timeBeforeBuildingApp.Add(-1 * time.Millisecond)) {
		t.Errorf("*** the app instance ULID time should not be before the time that the app was created: %v is not before %v",
			ulid.Time(appInstanceID.ULID().Time()),
			timeBeforeBuildingApp.Add(-1*time.Millisecond),
		)
	}

	checkConstructorsAreRegistered(t, app,
		ProvideBar,
		ProvideBaz,
		fooID.ProvideFooID,
		ProvidePasswordLogin,
		ProvideMFALogin,
		GroupMFALogin,
		GroupPasswordLogin,
	)

	checkFuncsAreRegistered(t, app,
		InvokePrintBaz,
		InvokePrintBar,
		fooID.InvokeLogInstanceID,
		fooID.InvokeLogSelf,
		InvokeLogin,
		GatherLogins,
	)

}

func checkConstructorsAreRegistered(t *testing.T, app fxapp.App, constructors ...interface{}) {
Loop:
	for _, c := range constructors {
		for _, t := range app.ConstructorTypes() {
			if t == reflect.TypeOf(c) {
				continue Loop
			}
		}
		t.Errorf("*** constructor was not registered: %v", reflect.TypeOf(c))
	}
}

func checkFuncsAreRegistered(t *testing.T, app fxapp.App, funcs ...interface{}) {
Loop:
	for _, f := range funcs {
		for _, t := range app.FuncTypes() {
			if t == reflect.TypeOf(f) {
				continue Loop
			}
		}
		t.Errorf("*** func was not registered: %v", reflect.TypeOf(f))
	}
}

func TestRunningApp(t *testing.T) {
	app, err := fxapp.NewAppBuilder(newDesc("foo", "0.1.0")).
		Invoke(
			// app InstanceID is automatically provided
			func(instanceID fxapp.InstanceID) {
				t.Logf("app instance ID: %v", instanceID)
			},
			// app Desc is automatically provided
			func(desc fxapp.Desc) {
				t.Logf("app desc: %v", desc)
			},
			// trigger shutdown
			func(lc fx.Lifecycle, shutdowner fx.Shutdowner) {
				lc.Append(fx.Hook{
					OnStart: func(context.Context) error {
						return shutdowner.Shutdown()
					},
				})
			},
		).
		Build()

	if err != nil {
		t.Fatalf("*** app failed to build: %v", err)
	}

	err = app.Run()
	if err != nil {
		t.Errorf("*** app failed to run: %v", err)
	}

	t.Logf("stop signal: %v", <-app.Done())
}

func TestAppLifeCycleSignals(t *testing.T) {
	app, err := fxapp.NewAppBuilder(newDesc("foo", "0.1.0")).
		Invoke(
			// trigger shutdown
			func(lc fx.Lifecycle, shutdowner fx.Shutdowner) {
				lc.Append(fx.Hook{
					OnStart: func(context.Context) error {
						return shutdowner.Shutdown()
					},
				})
			},
		).
		Build()

	if err != nil {
		t.Fatalf("*** app failed to build: %v", err)
	}

	// When the app is run async
	go func() {
		err := app.Run()
		if err != nil {
			t.Errorf("*** app failed to run: %v", err)
		}
	}()

	var lifecycleEvents []string
	<-app.Starting()
	lifecycleEvents = append(lifecycleEvents, "starting")
	t.Log("app is starting")

	<-app.Started()
	lifecycleEvents = append(lifecycleEvents, "started")
	t.Log("app has started")

	stopSignal := <-app.Stopping()
	lifecycleEvents = append(lifecycleEvents, "stopping")
	t.Logf("app is stopping: %v", stopSignal)

	doneDignal := <-app.Done()
	if doneDignal != stopSignal {
		t.Errorf("*** stop and done signals should be the same: %v : %v", stopSignal, doneDignal)
	}
	lifecycleEvents = append(lifecycleEvents, "stopped")
	t.Logf("app is stopped: %v", doneDignal)

	t.Logf("lifecycleEvents: %v", lifecycleEvents)
	if len(lifecycleEvents) != 4 {
		t.Errorf("*** lifecycle event count should be 4 but was : %d", len(lifecycleEvents))
	}
}

// When the application is run, the registered functions are invoked in the order that they are registered.
func TestFuncInvokeOrder(t *testing.T) {
	var funcInvokes []string
	var funcs []interface{}
	for i := 0; i < 10; i++ {
		ii := i
		funcs = append(funcs, func() {
			funcInvokes = append(funcInvokes, fmt.Sprintf("%d", ii))
		})
	}

	app, err := fxapp.NewAppBuilder(newDesc("foo", "0.1.0")).
		Invoke(funcs...).
		Invoke(
			func(lc fx.Lifecycle, shutdowner fx.Shutdowner) {
				funcInvokes = append(funcInvokes, "shutdown")
				lc.Append(fx.Hook{
					OnStart: func(context.Context) error {
						return shutdowner.Shutdown()
					},
				})
			},
		).
		Build()

	// functions are invoked when the app is built
	t.Logf("funcInvokes: %v", funcInvokes)
	if len(funcInvokes) != 11 {
		t.Errorf("*** the number of invoked functions should be 11 but was: %v", len(funcInvokes))
	}
	if funcInvokes[10] != "shutdown" {
		t.Errorf("*** the last func to be invoked should have been `shutdown`, but was %q", funcInvokes[10])
	}
	for i := 0; i < 10; i++ {
		funcs = append(funcs, func() {
			if funcInvokes[i] != fmt.Sprintf("%d", i) {
				t.Errorf("*** func[%d] invoked out of expected order: %v", i, funcInvokes[i])
			}
		})
	}

	if err != nil {
		t.Fatalf("*** app failed to build: %v", err)
	}

	err = app.Run()
	if err != nil {
		t.Errorf("*** app failed to run: %v", err)
	}
}

// error handlers can be registered with the application. They are executed on function invocation failures.
func TestFuncErrorHandling(t *testing.T) {
	funcInvocations := make(map[int]time.Time)
	var errHandleCount uint
	app, err := fxapp.NewAppBuilder(newDesc("foo", "0.1.0")).
		Invoke(
			func() error {
				funcInvocations[1] = time.Now()
				return nil
			},
			func() error {
				funcInvocations[2] = time.Now()
				return errors.New("func 2 failed")
			},
			func() error {
				funcInvocations[3] = time.Now()
				return errors.New("func 3 failed")
			}).
		HandleInvokeError(
			func(err error) {
				t.Logf("handler 1 received error: %v", err)
				errHandleCount++
			},
			func(err error) {
				t.Logf("handler 2 received error: %v", err)
				errHandleCount++
			},
		).
		Build()

	if app != nil {
		t.Error("*** app should be nil because it failed to build")
	}

	t.Logf("err: %v", err)
	if err == nil {
		t.Error("*** app should have failed to build")
	}

	t.Logf("funcInvocations: %v", funcInvocations)
	if funcInvocations[1].IsZero() {
		t.Error("*** func 1 should have ran")
	}
	if funcInvocations[2].IsZero() {
		t.Error("*** func 2 should have ran")
	}
	if !funcInvocations[3].IsZero() {
		t.Error("*** func 3 should not have run because func 2 should have failed before")
	}
	if !funcInvocations[1].Before(funcInvocations[2]) {
		t.Error("*** func 1 should have run before func 2")
	}

	if errHandleCount != 2 {
		t.Errorf("not all error handlers were invoked: %d", errHandleCount)
	}
}

// app default start and stop timeout is 15 sec
func TestAppDefaultStartStopTimeouts(t *testing.T) {
	t.Fatal("TODO")
}

// app can populate targets with values from the dependency injection container
func TestPopulate(t *testing.T) {
	t.Fatal("TODO")
}

// By default, the app logs to stderr. However, an alternative writer can be provided for logging when the app is being built.
func TestAppLogWriter(t *testing.T) {
	t.Fatal("TODO")
}
