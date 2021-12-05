package bn256

import (
	"testing"

	"bytes"
	"crypto/rand"

	"math/big"

	"golang.org/x/crypto/bn256"
)

func TestG1(t *testing.T) {
	k, Ga, err := RandomG1(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	ma := Ga.Marshal()

	Gb := new(bn256.G1).ScalarBaseMult(k)
	mb := Gb.Marshal()

	if !bytes.Equal(ma, mb) {
		t.Fatal("bytes are different")
	}
}

func TestG1Marshal(t *testing.T) {
	_, Ga, err := RandomG1(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	ma := Ga.Marshal()

	Gb := new(G1)
	_, err = Gb.Unmarshal(ma)
	if err != nil {
		t.Fatal(err)
	}
	mb := Gb.Marshal()

	if !bytes.Equal(ma, mb) {
		t.Fatal("bytes are different")
	}
}

func TestG2(t *testing.T) {
	k, Ga, err := RandomG2(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	ma := Ga.Marshal()

	Gb := new(bn256.G2).ScalarBaseMult(k)
	mb := Gb.Marshal()
	mb = append([]byte{0x01}, mb...)

	if !bytes.Equal(ma, mb) {
		t.Fatal("bytes are different")
	}
}

func TestG2Marshal(t *testing.T) {
	_, Ga, err := RandomG2(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	ma := Ga.Marshal()

	Gb := new(G2)
	_, err = Gb.Unmarshal(ma)
	if err != nil {
		t.Fatal(err)
	}
	mb := Gb.Marshal()

	if !bytes.Equal(ma, mb) {
		t.Fatal("bytes are different")
	}
}

func TestGT(t *testing.T) {
	k, Ga, err := RandomGT(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	ma := Ga.Marshal()

	Gb, ok := new(bn256.GT).Unmarshal((&GT{gfP12Gen}).Marshal())
	if !ok {
		t.Fatal("unmarshal not ok")
	}
	Gb.ScalarMult(Gb, k)
	mb := Gb.Marshal()

	if !bytes.Equal(ma, mb) {
		t.Fatal("bytes are different")
	}
}

func TestGTMarshal(t *testing.T) {
	_, Ga, err := RandomGT(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	ma := Ga.Marshal()

	Gb := new(GT)
	_, err = Gb.Unmarshal(ma)
	if err != nil {
		t.Fatal(err)
	}
	mb := Gb.Marshal()

	if !bytes.Equal(ma, mb) {
		t.Fatal("bytes are different")
	}
}

func TestBilinearity(t *testing.T) {
	for i := 0; i < 2; i++ {
		a, p1, _ := RandomG1(rand.Reader)
		b, p2, _ := RandomG2(rand.Reader)
		e1 := Pair(p1, p2)

		e2 := Pair(&G1{curveGen}, &G2{twistGen})
		e2.ScalarMult(e2, a)
		e2.ScalarMult(e2, b)

		if *e1.P != *e2.P {
			t.Fatalf("bad pairing result: %s", e1)
		}
	}
}

func TestTripartiteDiffieHellman(t *testing.T) {
	a, _ := rand.Int(rand.Reader, Order)
	b, _ := rand.Int(rand.Reader, Order)
	c, _ := rand.Int(rand.Reader, Order)

	pa, pb, pc := new(G1), new(G1), new(G1)
	qa, qb, qc := new(G2), new(G2), new(G2)

	pa.Unmarshal(new(G1).ScalarBaseMult(a).Marshal())
	qa.Unmarshal(new(G2).ScalarBaseMult(a).Marshal())
	pb.Unmarshal(new(G1).ScalarBaseMult(b).Marshal())
	qb.Unmarshal(new(G2).ScalarBaseMult(b).Marshal())
	pc.Unmarshal(new(G1).ScalarBaseMult(c).Marshal())
	qc.Unmarshal(new(G2).ScalarBaseMult(c).Marshal())

	k1 := Pair(pb, qc)
	k1.ScalarMult(k1, a)
	k1Bytes := k1.Marshal()

	k2 := Pair(pc, qa)
	k2.ScalarMult(k2, b)
	k2Bytes := k2.Marshal()

	k3 := Pair(pa, qb)
	k3.ScalarMult(k3, c)
	k3Bytes := k3.Marshal()

	if !bytes.Equal(k1Bytes, k2Bytes) || !bytes.Equal(k2Bytes, k3Bytes) {
		t.Errorf("keys didn'T agree")
	}
}

func TestGfPSqrt(t *testing.T) {
	// test square root function in gfP
	s, err := rand.Int(rand.Reader, p)

	ss := new(big.Int).Mul(s, s)
	ss.Mod(ss, p)
	sMinus := new(big.Int).Neg(s)
	sMinus.Mod(sMinus, p)

	a, aa, aaSqrt := &gfP{}, &gfP{}, &gfP{}
	a.SetInt(s)
	gfpMul(aa, a, a)

	aaSqrt.Sqrt(aa)
	ssSqrt, err := aaSqrt.ToInt()
	if err != nil {
		t.Errorf("convertion failed: %v", err)
	}

	if ssSqrt.Cmp(s) != 0 && ssSqrt.Cmp(sMinus) != 0 {
		t.Errorf("wrong result for GfP")
	}

	// test square root function in gfP2
	a2 := &gfP2{*a, *aa}

	a2a2, a2a2Sqrt, a2Minus := &gfP2{}, &gfP2{}, &gfP2{}
	a2a2.Mul(a2, a2)
	a2Minus.Neg(a2)

	_, err = a2a2Sqrt.Sqrt(a2a2)

	if a2.String() != a2a2Sqrt.String() && a2Minus.String() != a2a2Sqrt.String() {
		t.Errorf("wrong result for GfP2")
	}
}

func TestHashToG1(t *testing.T) {
	s := "foo bar"
	_, err := HashG1(s)
	if err != nil {
		t.Errorf("hashing failed: %v", err)
	}
}

func TestHashToG2(t *testing.T) {
	s := "foo bar"
	_, err := HashG2(s)
	if err != nil {
		t.Errorf("hashing failed: %v", err)
	}
}

func BenchmarkG1(b *testing.B) {
	x, _ := rand.Int(rand.Reader, Order)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		new(G1).ScalarBaseMult(x)
	}
}

func BenchmarkG2(b *testing.B) {
	x, _ := rand.Int(rand.Reader, Order)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		new(G2).ScalarBaseMult(x)
	}
}

func BenchmarkGT(b *testing.B) {
	x, _ := rand.Int(rand.Reader, Order)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		new(GT).ScalarBaseMult(x)
	}
}

func BenchmarkPairing(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Pair(&G1{curveGen}, &G2{twistGen})
	}
}
