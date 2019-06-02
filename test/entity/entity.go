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
	"github.com/oklog/ulid"
	"golang.org/x/crypto/nacl/sign"
)

// Domain is used to represent an Entity domain.
// Domains can form a hierarchy, i.e., all Domains have a parent Domain, except for the root Domain.
type Domain struct {
	DomainID
	Name
	ParentID DomainID
	// BoxPublicKey is used to exchange encrypted messages with the Domain
	BoxPublicKey  BoxPublicKey
	SignPublicKey SignPublicKey
}

// IsRoot returns true of ParentID is a zero value
func (d Domain) IsRoot() bool {
	return d.IsZero()
}

// DomainID is used as the Entity's domain ID
type DomainID ulid.ULID

// IsZero returns true of the DomainID has no value, i.e., is the zero value
func (d DomainID) IsZero() bool {
	return isULIDZero(ulid.ULID(d))
}

// Entity represents any entity instance for a given Class
type Entity struct {
	// DomainID for the Entity owner
	DomainID
	ID
	// Name is the unique Entity name within
	Name
	// BoxPublicKey is used to exchange encrypted messages
	BoxPublicKey BoxPublicKey
}

// ID is used as the Entity identifier.
// The ID timestamp correlates to when the Entity was created.
type ID ulid.ULID

// Name is used to assign a user friendly name for an Entity
type Name string

// BoxPublicKey represents an Entity's public key for public-key cryptography that is interoperable with NaCl: https://nacl.cr.yp.to/box.html.
//
// See: https://godoc.org/golang.org/x/crypto/nacl/box
type BoxPublicKey *[32]byte

// BoxPrivateKey represents an Entity's crypto private key for public-key cryptography that is interoperable with NaCl: https://nacl.cr.yp.to/box.html.
type BoxPrivateKey *[32]byte

// SignPublicKey is the public key used to verify signatures
//
// See: https://godoc.org/golang.org/x/crypto/nacl/sign
type SignPublicKey *[32]byte

// SignPrivateKey represents an Entity's crypto private key for signing messages.
//
// See: https://godoc.org/golang.org/x/crypto/nacl/sign
type SignPrivateKey *[64]byte

// SignedULID is a signed ULID
type SignedULID []byte

// SignULID signs the specified ULID
func SignULID(id ulid.ULID, key SignPrivateKey) SignedULID {
	return SignedULID(sign.Sign(nil, id[:], (*[64]byte)(key)))
}

// Open verifies the signed ULID using the specified public key
func (su SignedULID) Open(key SignPublicKey) (id ulid.ULID, ok bool) {
	msg, ok := sign.Open(nil, su[:], (*[32]byte)(key))
	if ok {
		if err := id.UnmarshalBinary(msg); err != nil {
			panic(err)
		}
	}
	return
}

func isULIDZero(uid ulid.ULID) bool {
	var zero ulid.ULID
	return uid == zero
}
