package kvstore

import (
	"fmt"
	"strconv"
	"testing"
)

var (
	testId            = "0"
	testPrefix        = "TS"
	testMetadata      = [...]byte{'T', 'E', 'S', 'T'}
	testNumLogEntries = 10
)

type testRegistry struct {
	t *testing.T
}

func (*testRegistry) Prefix() string {
	return testPrefix
}

func (r *testRegistry) NewReplay(id string) ReplayHandler {
	if testId != id {
		r.t.Fatalf("NewReply incorrect ID: Expected: %s Actual: %s", testId, id)
	}
	return &testReply{t: r.t}
}

type testReply struct {
	t *testing.T
}

func (r *testReply) Metadata(data []byte) error {
	if len(testMetadata) != len(data) {
		r.t.Fatalf("TestReply metadata length incorrect: Expected: %d Actual: %d", len(testMetadata), len(data))
	}

	for i := 0; i < len(testMetadata); i++ {
		if testMetadata[i] != data[i] {
			r.t.Fatalf("TestReply data mistmatch: Expected: %s Actual: %s", string(testMetadata[:]), string(data))
		}
	}

	return nil
}

func (r *testReply) Entry(t uint32, data []byte) error {
	if t > uint32(testNumLogEntries) {
		r.t.Fatalf("TestEntry type invalid: %d", t)
	}

	val, err := strconv.Atoi(string(data))
	if err != nil {
		return err
	}
	if int(t) != val {
		r.t.Fatalf("TestEntry invalid data value: Expected: %d Actual: %d", t, val)
	}

	return nil
}

func (*testReply) Done() error {
	return nil
}

func TestStore(t *testing.T) {
	store, err := Open("testing.db", false)
	if err != nil {
		t.Errorf("Failed to open testing.db: Error: %s", err)
		t.Fail()
	}

	defer store.Close()

	registry := testRegistry{t: t}
	store.Register([]Registry{&registry})

	// Create a new key
	{
		ledger, err := store.NewKey(store.MakeKey(&registry, testId), testMetadata[:])
		if err != nil {
			t.Errorf("Failed to create new ledger key %s: Error: %s", testId, err)
		}

		for i := 0; i < testNumLogEntries; i++ {

			if err := ledger.Log(uint32(i), []byte(fmt.Sprintf("%d", i))); err != nil {
				t.Errorf("Failed to log ledger entry %d: Error: %s", i, err)
			}
		}

		ledger.Close()
	}

	// Open an existing key
	{
		ledger, err := store.OpenKey(store.MakeKey(&registry, testId), false)
		if err != nil {
			t.Errorf("Failed to open existing ledger key %s: Error: %s", testId, err)
		}

		for i := 0; i < testNumLogEntries; i++ {

			if err := ledger.Log(uint32(i), []byte(fmt.Sprintf("%d", i))); err != nil {
				t.Errorf("Failed to log ledger entry %d: Error: %s", i, err)
			}
		}

		ledger.Close()
	}

	// Run the ledger replay
	{
		if err := store.Replay(); err != nil {
			t.Errorf("Failed to run reply: Error: %s", err)
		}
	}
}
