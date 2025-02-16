/*
Copyright 2015 Gravitational, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package native

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	mathrand "math/rand"
	"sync"
	"testing"
	"time"

	"github.com/gravitational/trace"
	"github.com/sirupsen/logrus"

	"github.com/gravitational/teleport"
	"github.com/gravitational/teleport/api/constants"
	"github.com/gravitational/teleport/api/utils/keys"
)

var log = logrus.WithFields(logrus.Fields{
	trace.Component: teleport.ComponentKeyGen,
})

// precomputedKeys is a queue of cached keys ready for usage.
var precomputedKeys = make(chan *rsa.PrivateKey, 25)

// startPrecomputeOnce is used to start the background task that precomputes key pairs.
var startPrecomputeOnce sync.Once

// GenerateKeyPair generates a new RSA key pair.
func GenerateKeyPair() ([]byte, []byte, error) {
	priv, err := GeneratePrivateKey()
	if err != nil {
		return nil, nil, trace.Wrap(err)
	}
	return priv.PrivateKeyPEM(), priv.MarshalSSHPublicKey(), nil
}

// GeneratePrivateKey generates a new RSA private key.
func GeneratePrivateKey() (*keys.PrivateKey, error) {
	rsaKey, err := getOrGenerateRSAPrivateKey()
	if err != nil {
		return nil, trace.Wrap(err)
	}

	// We encode the private key in PKCS #1, ASN.1 DER form
	// instead of PKCS #8 to maintain compatibility with some
	// third party clients.
	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:    keys.PKCS1PrivateKeyType,
		Headers: nil,
		Bytes:   x509.MarshalPKCS1PrivateKey(rsaKey),
	})

	return keys.NewPrivateKey(rsaKey, keyPEM)
}

func GenerateRSAPrivateKey() (*rsa.PrivateKey, error) {
	return getOrGenerateRSAPrivateKey()
}

func getOrGenerateRSAPrivateKey() (*rsa.PrivateKey, error) {
	select {
	case k := <-precomputedKeys:
		return k, nil
	default:
		rsaKeyPair, err := generateRSAPrivateKey()
		if err != nil {
			return nil, err
		}
		return rsaKeyPair, nil
	}
}

func generateRSAPrivateKey() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, constants.RSAKeySize)
}

func precomputeKeys() {
	const backoff = time.Second * 30
	for {
		rsaPrivateKey, err := generateRSAPrivateKey()
		if err != nil {
			log.WithError(err).Errorf("Failed to precompute key pair, retrying in %s (this might be a bug).", backoff)
			time.Sleep(backoff)
		}

		precomputedKeys <- rsaPrivateKey
	}
}

func precomputeTestKeys() {
	testKeys, err := generateTestKeys()
	if err != nil {
		// Use only in tests. Safe to panic.
		panic(err)
	}

	for {
		// Shuffle keys to reduce potential multiple usage of the
		// same key in one test.
		mathrand.Shuffle(len(testKeys), func(i, j int) {
			testKeys[i], testKeys[j] = testKeys[j], testKeys[i]
		})

		for _, k := range testKeys {
			precomputedKeys <- k
		}
	}
}

// testKeysNumber is the number of RSA keys generated in tests.
const testKeysNumber = 25

func generateTestKeys() ([]*rsa.PrivateKey, error) {
	privateKeys := make([]*rsa.PrivateKey, 0, testKeysNumber)
	keysChan := make(chan *rsa.PrivateKey)
	errC := make(chan error)

	go func() {
		for i := 0; i < testKeysNumber; i++ {
			// Generate each key in a separate goroutine to take advantage of
			// multiple cores if possible.
			go func() {
				private, err := generateRSAPrivateKey()
				if err != nil {
					errC <- trace.Wrap(err)
					return
				}
				keysChan <- private
			}()
		}
	}()

	for i := 0; i < testKeysNumber; i++ {
		select {
		case err := <-errC:
			return nil, trace.Wrap(err)
		case privKey := <-keysChan:
			privateKeys = append(privateKeys, privKey)
		}
	}

	return privateKeys, nil
}

// PrecomputeKeys sets this package into a mode where a small backlog of keys are
// computed in advance. This should only be enabled if large spikes in key computation
// are expected (e.g. in auth/proxy services). Safe to double-call.
func PrecomputeKeys() {
	startPrecomputeOnce.Do(func() {
		go precomputeKeys()
	})
}

// PrecomputeTestKeys generates RSA keys and reuse them to reduce CPU usage. This method should
// only be in tests. Safe to call multiple times.
// This function takes *testing.M, so is only can be used from TestMain in tests.
func PrecomputeTestKeys(_ *testing.M) {
	startPrecomputeOnce.Do(func() {
		go precomputeTestKeys()
	})
}
