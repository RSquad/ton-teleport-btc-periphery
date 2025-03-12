package frost

import (
	"encoding/hex"
	"fmt"
	"testing"
)

func GetIdentifierFromHex(hexStr string) []byte {
	data, err := hex.DecodeString(hexStr)
	if err != nil {
		panic(err)
	}
	return data
}

func TestHex(t *testing.T) {
	hexStr := "46447381F0"
	data, err := hex.DecodeString(hexStr)
	if err != nil {
		panic(err)
	}
	fmt.Printf("% x", data)
	fmt.Println()
}

func TestGetIdentifier(t *testing.T) {
	data := GetIdentifier(1)
	fmt.Printf("%v", data)
	fmt.Println()

	data = GetIdentifier(2)
	fmt.Printf("%v", data)
	fmt.Println()

	data = GetIdentifier(3)
	fmt.Printf("%v", data)
	fmt.Println()
}

func TestGetIdentifierFromHex(t *testing.T) {
	strTestHex := "ffeeddccbbaa00112233445566778899ffeeddccbbaa00112233445566778899"
	bId := GetIdentifierFromHex(strTestHex)

	fmt.Printf("%x", bId)
	fmt.Println()
}

func TestDkgPart1(t *testing.T) {
	fmt.Printf("%s\n", t.Name())
	identifier := "900924ca1a6d37bd613419f55038d4e210c4e347cf9a8f128181c823684a212f"
	idVec, _ := hex.DecodeString(identifier)
	pkg, sp, err := DkgPart1(idVec, 2, 3)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("Secret ptr %x", sp)
	fmt.Println()
	fmt.Printf("Package %x", pkg)
	fmt.Println()
}

// Test DkgPart1 + DkgPart2 + DkgPart3
func TestDKG(t *testing.T) {
	fmt.Printf("%s\n", t.Name())
	identifiers := [3]string{
		"87c848293689b356a7cf032b1c97d56955c0e1ba5d87ed36c4d6557520c3e0e6",
		"900924ca1a6d37bd613419f55038d4e210c4e347cf9a8f128181c823684a212f",
		"29573dedfaa8f3ac724687387289224816d16fa934b4f5fd6ebcffe55a1f0c28",
	}
	var maxSigners uint16 = 3
	var minSigners uint16 = 2
	r1Secrets := make(map[Identifier]uintptr)
	r2Secrets := make(map[Identifier]uintptr)
	receivedR1Packages := make(map[Identifier]map[Identifier]Package)
	receivedR2Packages := make(map[Identifier]map[Identifier]Package)

	for i := uint16(0); i < maxSigners; i++ {
		sender, _ := DecodeIdentifier(identifiers[i])
		pkg, sp, err := DkgPart1(sender[:], minSigners, maxSigners)
		if err != nil {
			t.Error(err)
		}
		r1Secrets[*sender] = sp
		for j := uint16(0); j < maxSigners; j++ {
			if i != j {
				receiver, _ := DecodeIdentifier(identifiers[j])
				incomePackages, ok := receivedR1Packages[*receiver]
				if !ok {
					incomePackages = make(map[Identifier]Package)
				}
				incomePackages[*sender] = Package{pkg}
				receivedR1Packages[*receiver] = incomePackages
			}
		}
	}

	for i := uint16(0); i < maxSigners; i++ {
		sender, _ := DecodeIdentifier(identifiers[i])
		secret := r1Secrets[*sender]
		delete(r1Secrets, *sender)
		r2Packages, r2s, err := DkgPart2(secret, receivedR1Packages[*sender])
		if err != nil {
			t.Error(err)
		}
		fmt.Printf("round2Packages %x\n", r2Packages)
		r2Secrets[*sender] = r2s
		for receiver, r2package := range r2Packages {
			incomePackages, ok := receivedR2Packages[receiver]
			if !ok {
				incomePackages = make(map[Identifier]Package)
			}
			incomePackages[*sender] = r2package
			receivedR2Packages[receiver] = incomePackages
		}
	}

	keyPackages := make(map[Identifier]Package)
	pubkeyPackages := make(map[Identifier]Package)
	for i := uint16(0); i < maxSigners; i++ {
		sender, _ := DecodeIdentifier(identifiers[i])
		keyPackage, publicKeyPackage, err := DkgPart3(
			r2Secrets[*sender],
			receivedR1Packages[*sender],
			receivedR2Packages[*sender],
		)
		if err != nil {
			t.Error(err)
		}
		keyPackages[*sender] = Package{keyPackage}
		pubkeyPackages[*sender] = Package{publicKeyPackage}
	}
	fmt.Printf("PublicKeyPackages %x\nKeyPackages %x\n", pubkeyPackages, keyPackages)

	commitments := make(map[Identifier]Package)
	nonces := make(map[Identifier]Package)
	for i := uint16(0); i < maxSigners; i++ {
		sender, _ := DecodeIdentifier(identifiers[i])
		keyPackage := keyPackages[*sender]
		// fmt.Printf("keyPackage(%d) %x\n", len(keyPackage.buf), keyPackage)
		nonce, commitment, err := Commit(keyPackage)
		if err != nil {
			t.Error(err)
		}
		// fmt.Printf("Nonces %x\nCommitments %x\n", nonce, commitment)
		nonces[*sender] = Package{nonce}
		commitments[*sender] = Package{commitment}
	}

	strTestHex := "ffeeddccbbaa00112233445566778899ffeeddccbbaa00112233445566778899"
	message := GetIdentifierFromHex(strTestHex)

	signatureShares := make(map[Identifier]Package)
	for i := uint16(0); i < maxSigners; i++ {
		sender, _ := DecodeIdentifier(identifiers[i])

		signatureShare, err := SignWithTweak(
			// merkleRoot,
			keyPackages[*sender],
			message,
			commitments,
			nonces[*sender],
		)
		if err != nil {
			t.Error(err)
		}
		fmt.Printf("SignatureShares %x\n", signatureShare)
		signatureShares[*sender] = Package{signatureShare}
	}

	sender, _ := DecodeIdentifier(identifiers[0])
	groupSignature, err := AggregateWithTweak(
		message,
		commitments,
		signatureShares,
		pubkeyPackages[*sender],
	)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("groupSignature %x\n", groupSignature)

	res, err := Verify(
		pubkeyPackages[*sender],
		message,
		groupSignature,
	)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("Verify -> %t\n", res)
}
