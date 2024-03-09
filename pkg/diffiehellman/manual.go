package diffiehellman

import (
	"crypto/rand"
	"errors"
	"log"
	"math/big"
)

const (
	MANUAL_BIT_SIZE    = 64
	MAX_PRIMITIVE_ROOT = 100
)

// Helper function to find the smallest primitive root modulo a prime number
func getSmallestPrimitiveRoot(base *big.Int) (int, error) {
	// Calculate the Euler's totient function of the base
	// Since the base is the prime number, the Euler's totient function is base-1
	// phi = base - 1
	phi := new(big.Int).Sub(base, big.NewInt(1))
	// Calculate the unique prime factors of the Euler's totient function
	primeFactors := primeFactors(new(big.Int).Set(phi))
	log.Println("primeFactors: ", primeFactors)
	// Iterate through numbers from 2 to base-1
	for i := 2; i < MAX_PRIMITIVE_ROOT; i++ {
		// Check if i is a primitive root modulo base
		if isPrimitiveRoot(big.NewInt(int64(i)), base, phi, primeFactors) {
			return i, nil
		}
	}
	return 0, errors.New("no primitive root found for the base")
}

// Helper function to calculate the prime factors of a number, and return only the unique ones
// Prime Factors - a combination of numbers, which when multiplied together, give the original number
func primeFactors(n *big.Int) []*big.Int {
	// Create a slice to store the prime factors
	factors := []*big.Int{}
	two := big.NewInt(2)
	// Iterate through all even numbers
	for new(big.Int).Mod(n, two).Cmp(big.NewInt(0)) == 0 {
		// Add 2 to the set of prime factors
		factors = append(factors, two)
		n.Div(n, two)
	}
	// Iterate through all odd numbers
	for i := big.NewInt(3); new(big.Int).Mul(i, i).Cmp(n) <= 0; i.Add(i, two) {
		for new(big.Int).Mod(n, i).Cmp(big.NewInt(0)) == 0 {
			factors = append(factors, new(big.Int).Set(i))
			n.Div(n, i)
		}
	}
	if n.Cmp(two) > 0 {
		factors = append(factors, n)
	}
	return factors
}

// Helper function to check if a number is a primitive root modulo base
func isPrimitiveRoot(a *big.Int, base *big.Int, phi *big.Int, primeFactors []*big.Int) bool {
	// Iterate through all prime factors of the Euler's totient function
	for _, prime := range primeFactors {
		// Check if a^(phi/prime) != 1 (mod base)
		if new(big.Int).Exp(a, new(big.Int).Div(phi, prime), base).Cmp(big.NewInt(1)) == 0 {
			return false
		}
	}
	return true
}

// A manual implementation of the base secrets generation for the Diffie-Hellman key exchange
func GenerateManualBaseSecrets() (*big.Int, int, error) {
	// Generate a base prime number, which will be used as a modulus
	// The number will be approixmately 600+ digits long
	p, err := rand.Prime(rand.Reader, BIT_SIZE)
	if err != nil {
		return &big.Int{}, 0, err
	}
	log.Println("p: ", p.String())

	// Generate a primitive root modulo of the prime number
	// The primitive root modulo is a number that is coprime to the prime number
	// and has a multiplicative order modulo p
	// The multiplicative order of a number a modulo p is the smallest positive integer k
	// such that a^k = 1 (mod p)
	g, err := getSmallestPrimitiveRoot(p)
	if err != nil {
		return &big.Int{}, 0, err
	}

	return p, g, nil
}
