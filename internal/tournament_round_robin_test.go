package internal

import "testing"

func TestRoundRobinCircleIndex(t *testing.T) {
	length := 14

	r1 := roundRobinCircleIndex(0, length, 0)
	r2 := roundRobinCircleIndex(0, length, 7)
	if r1 != 0 || r2 != 0 {
		t.Fatal("Index 0 did not stay fixed")
	}

	r1 = roundRobinCircleIndex(1, length, 0)
	r2 = roundRobinCircleIndex(5, length, 0)
	if r1 != 1 || r2 != 5 {
		t.Fatal("First round rotation did not preserve the original index")
	}

	r1 = roundRobinCircleIndex(1, length, 1)
	r2 = roundRobinCircleIndex(length-1, length, 1)
	if r1 != 2 || r2 != 1 {
		t.Fatal("Second round index was not rotated by one")
	}

	r1 = roundRobinCircleIndex(1, length, length-2)
	if r1 != length-1 {
		t.Fatal("The last round did not rotate index 1 to the last index")
	}
}
