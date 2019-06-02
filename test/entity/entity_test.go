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

package entity

import (
	"crypto/rand"
	"github.com/oklog/ulid"
	"golang.org/x/crypto/nacl/sign"
	"testing"
	"time"
)

func TestName(t *testing.T) {
	name := Name("alfio")
	t.Log(name)
}

var ULIDZero ulid.ULID

func TestULIDZeroValue(t *testing.T) {
	t.Parallel()

	var zero ulid.ULID
	t.Logf("ULID zero value: %v", zero)

	if zero != ULIDZero {
		t.Errorf("zero == ULIDZero assertion failed: %v != %v", zero, ULIDZero)
	}

	uid := ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader)
	if uid == ULIDZero {
		t.Errorf("ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader) != ULIDZero assertion failed: %v != %v", uid, ULIDZero)
	}
}

func TestSignULID(t *testing.T) {
	pubKey, privKey, err := sign.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	id := ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader)

	signedULID := SignULID(id, SignPrivateKey(privKey))
	openedULID, ok := signedULID.Open(SignPublicKey(pubKey))
	if !ok {
		t.Fatalf("failed to open signed message")
	}
	if openedULID != id {
		t.Fatalf("ids did not match : %s != %s", id, openedULID)
	}

	//signedMsg := sign.Sign(nil, id[:], privKey)
	//var temp ulid.ULID
	//temp.UnmarshalBinary(signedMsg[sign.Overhead:])
	//t.Logf("id = %s, signedMsg = %s, len(signedMsg) = %d", id, temp, len(signedMsg))
	//
	//msg, ok := sign.Open(nil, signedMsg, pubKey)
	//if !ok {
	//	t.Errorf("failed to open signed message")
	//}
	//var id2 ulid.ULID
	//id2.UnmarshalBinary(msg)
	//t.Logf("id2 = %s", id2)
	//if id != id2 {
	//	t.Fatalf("ids did not match : %s != %s", id, id2)
	//}

}
