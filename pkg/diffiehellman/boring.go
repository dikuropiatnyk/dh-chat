package diffiehellman

import (
	"crypto/rand"
	"math/big"
)

const (
	BIT_SIZE  = 2048
	GENERATOR = 2
	KEY_SIZE  = 256
)

func GenerateBaseSecrets() (*big.Int, *big.Int, error) {
	// Generate a base prime number (q), which will be used to get a modulus
	// The number will be approixmately 600+ digits long
	q, err := rand.Prime(rand.Reader, BIT_SIZE)
	if err != nil {
		return &big.Int{}, &big.Int{}, err
	}
	// Create a "strong" prime number, which is a prime number of the form 2q+1 - this will be the modulus
	p := new(big.Int).Mul(big.NewInt(2), q)
	// With such approach, p will be guaranteed to be a huge prime number,
	// g (generator) can be 2, because such p values will always
	// have primitive root modulo of {1, 2, q, 2q}
	// More details - https://crypto.stackexchange.com/a/829
	return p, big.NewInt(GENERATOR), nil
}

func GeneratePrivateSalt(p *big.Int) (*big.Int, error) {
	shift := big.NewInt(1)
	// Generate a private secret, which is a random number between 1 and p-1
	// The number will be used to generate a public key
	// The private secret will be used only by the client
	privateSecret, err := rand.Int(rand.Reader, new(big.Int).Sub(p, shift))
	if err != nil {
		return &big.Int{}, err
	}
	return new(big.Int).Add(privateSecret, shift), nil
}

func GeneratePublicSalt(p *big.Int, g *big.Int, privateSalt *big.Int) *big.Int {
	// Generate a public key, which is a number g^privateSalt mod p
	// The number will be used to generate a shared secret
	// The public key will be sent to the server
	return new(big.Int).Exp(g, privateSalt, p)
}

func GenerateSymmetricKey(p *big.Int, publicSalt *big.Int, privateSalt *big.Int) *big.Int {
	// Generate a symmetric key, which is a number publicSalt^privateSalt mod p
	// The number will be used to encrypt and decrypt messages
	return new(big.Int).Exp(publicSalt, privateSalt, p)
}
