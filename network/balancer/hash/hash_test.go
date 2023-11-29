package hash

import (
	"context"
	"fmt"
	"hash/crc32"
	"math"
	"math/rand"
	"sort"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/walleframe/walle/network"
	"github.com/walleframe/walle/network/balancer"
	"github.com/walleframe/walle/network/discovery"
	"github.com/walleframe/walle/testpkg/mock_discovery"
	"github.com/walleframe/walle/testpkg/mock_network"
)

const (
	balanceTestTimes = 8
)

func TestBalance(t *testing.T) {
	// prepare test
	mctl := gomock.NewController(t)
	entries := make([]discovery.Entry, 0, balanceTestTimes)
	state := make([]discovery.EntryState, balanceTestTimes)
	for i := 0; i < balanceTestTimes; i++ {
		k := i
		_ = k
		state[i] = discovery.EntryStateOnline
		client := mock_network.NewMockClient(mctl)
		entry := mock_discovery.NewMockEntry(mctl)
		entry.EXPECT().String().AnyTimes().Return(fmt.Sprintf("10.0.1.%d:8080", i))
		entry.EXPECT().Client().AnyTimes().Return(network.Client(client))
		entry.EXPECT().State().AnyTimes().DoAndReturn(func() discovery.EntryState {
			return state[k]
		})
		entry.EXPECT().Metadata("id").AnyTimes().Return(fmt.Sprint(i))
		entries = append(entries, entry)
	}
	// start
	builder := NewBalancer()
	picker := builder.Build(entries)
	assert.NotNil(t, picker, "new balancer picker")

	ctx := context.Background()
	// pick
	for i := 0; i < balanceTestTimes; i++ {
		pctx := balancer.WithBalanceIndex(ctx, int64(i))
		entry, err := picker.Pick(pctx)
		assert.Nil(t, err, "[%d]pick one not error", i)
		assert.NotNil(t, entry, "[%d]pick one entry", i)
		assert.NotNil(t, entry.Client(), "[%d]get entry client", i)
		assert.Equal(t, fmt.Sprint((i)%balanceTestTimes), entry.Metadata("id"), "[%d]compare id", i)
	}
	for k := range state {
		if k%2 == 0 {
			state[k] = discovery.EntryStateOffline
		}
	}
	for i := 0; i < balanceTestTimes; i++ {
		pctx := balancer.WithBalanceIndex(ctx, int64(i*2))
		entry, err := picker.Pick(pctx)
		assert.Nil(t, entry, "[%d] pick invalid")
		assert.NotNil(t, err, "[%d]pick one not error", i)
	}
}

type pair struct {
	v uint32
	f string
}

var generateCrc = func(n int, key, flag string) (ret []pair) {
	kn := uint32(crc32.ChecksumIEEE([]byte(key)))
	ks := uint32(uint64(1<<32) / uint64(n))
	//kf :=strconv.FormatUint(, 10)
	for i := 0; i < n; i++ {
		// v := crc32.ChecksumIEEE([]byte(fmt.Sprint(i) + key))
		// v := crc32.ChecksumIEEE([]byte(fmt.Sprintf("%03d#%s", i, key)))
		//v := crc32.ChecksumIEEE([]byte(fmt.Sprintf("%03d#%s", i, kf)))
		//v := crc32.ChecksumIEEE([]byte(fmt.Sprintf("%s#%d", key, i)))
		v := kn + uint32(i)*ks
		v = crc32.ChecksumIEEE([]byte(fmt.Sprintf("%d", v)))
		ret = append(ret, pair{
			f: flag,
			v: v,
		})
	}
	return
}

func makePairList(keyCount, itemCount int, key func(i int) string) []pair {
	list := make([]pair, 0, keyCount*itemCount)
	for i := 0; i < keyCount; i++ {
		list = append(list, generateCrc(itemCount, key(i), fmt.Sprintf("%c", 'A'+i))...)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].v < list[j].v
	})
	return list
}

func appendPairList(list []pair, itemCount int, key func() (key, flag string)) []pair {
	k, f := key()
	list = append(list, generateCrc(itemCount, k, f)...)
	sort.Slice(list, func(i, j int) bool {
		return list[i].v < list[j].v
	})
	return list
}

func makeRingArray(list []pair, arraySize int) (ring []string, collection map[string]int) {
	collection = make(map[string]int)
	ring = make([]string, arraySize)

	limit := (1 << 32) / arraySize

	begin := int(list[0].v / uint32(limit))
	for k := 0; k < int(begin); k++ {
		ring[k] = list[len(list)-1].f
		collection[list[len(list)-1].f]++
	}
	for k := 0; k < len(list)-1; k++ {
		up := list[k+1].v / uint32(limit)
		for i := begin; i < int(up); i++ {
			ring[i] = list[k].f
			collection[list[k].f]++
		}
		begin = int(up)
	}
	for k := list[len(list)-1].v / uint32(limit); k < 1024; k++ {
		ring[k] = list[len(list)-1].f
		collection[list[len(list)-1].f]++
	}

	return
}

func makeRingArrayFloat(list []pair, arraySize int) (ring []string, collection map[string]int) {
	collection = make(map[string]int)
	ring = make([]string, arraySize)

	limit := math.MaxUint32 / float64(arraySize)

	begin := int(list[0].v / uint32(limit))
	for k := 0; k < int(begin); k++ {
		ring[k] = list[len(list)-1].f
		collection[list[len(list)-1].f]++
	}
	for k := 0; k < len(list)-1; k++ {
		up := list[k+1].v / uint32(limit)
		for i := begin; i < int(up); i++ {
			ring[i] = list[k].f
			collection[list[k].f]++
		}
		begin = int(up)
	}
	for k := list[len(list)-1].v / uint32(limit); k < 1024; k++ {
		ring[k] = list[len(list)-1].f
		collection[list[len(list)-1].f]++
	}

	return
}

type diffrence struct {
	index int
	from  string
	to    string
}

func diffRing(from, to []string) (diff []diffrence) {
	for i := 0; i < len(from); i++ {
		if from[i] == to[i] {
			continue
		}
		diff = append(diff, diffrence{
			index: i,
			from:  from[i],
			to:    to[i],
		})
	}
	return
}

func TestHashRing(t *testing.T) {
	list1 := makePairList(3, 32, func(i int) string {
		return uuid.NewString()
	})
	ring1024, c1024 := makeRingArray(list1, 1024)
	ring2048, c2048 := makeRingArray(list1, 2048)
	t.Log(c1024, 1024/3)
	t.Log(c2048, 2048/3)

	list2 := appendPairList(list1, 32, func() (key string, flag string) {
		return uuid.NewString(), "-"
	})
	ring1024c, c1024c := makeRingArray(list2, 1024)
	ring2048c, c2048c := makeRingArray(list2, 2048)
	t.Log(c1024c, 1024/4)
	t.Log(c2048c, 2048/4)

	d1024 := diffRing(ring1024, ring1024c)
	d2048 := diffRing(ring2048, ring2048c)

	t.Log(d1024)
	//t.Log(d2048)

	t.Log("change-size:", len(d1024), float64(len(d1024))/1024*100.0)
	t.Log("change-size:", len(d2048), float64(len(d2048))/2048*100.0)
}

const (
	p3 = 1024.0 / 3
	p4 = 1024.0 / 4

	N = 1e3
)

func TestDiffenceCrcKey(t *testing.T) {

	calcPercent := func() (min3, max3, min4, max4, change float64) {
		list1 := makePairList(3, 32, func(i int) string {
			return uuid.NewString()
		})
		_, c1024 := makeRingArray(list1, 1024)
		min3 = math.MaxFloat64
		for _, v := range c1024 {
			p := float64(v) / p3 * 100
			if p > max3 {
				max3 = p
			}
			if p < min3 {
				min3 = p
			}
		}

		list2 := appendPairList(list1, 32, func() (key string, flag string) {
			return uuid.NewString(), "-"
		})
		_, c1024c := makeRingArray(list2, 1024)
		min4 = math.MaxFloat64
		for _, v := range c1024c {
			p := float64(v) / p4 * 100
			if p > max4 {
				max4 = p
			}
			if p < min4 {
				min4 = p
			}
		}
		change = float64(c1024c["-"]) / 1024 * 100
		return
	}
	genTest := func(mf func(n int, key, flag string) (ret []pair)) (rf func(t *testing.T)) {
		generateCrc = mf
		return func(t *testing.T) {
			min3, max3, min4, max4, change := calcPercent()
			for i := 0; i < N; i++ {
				a1, a2, a3, a4, a5 := calcPercent()
				if min3 > a1 {
					min3 = a1
				}
				if min4 > a3 {
					min4 = a3
				}
				if a2 > max3 {
					max3 = a2
				}
				if a4 > max4 {
					max4 = a4
				}
				change += a5
			}
			t.Logf("times:%8d 3(%2.2f, %2.2f) 4(%2.2f, %2.2f) change:%2.3f",
				int64(N+1),
				min3, max3, min4, max4,
				change/(N+1),
			)
		}

	}
	//

	// v1
	t.Run("%n+key", genTest(func(n int, key, flag string) (ret []pair) {
		for i := 0; i < n; i++ {
			v := crc32.ChecksumIEEE([]byte(fmt.Sprint(i) + "#" + key))
			ret = append(ret, pair{
				f: flag,
				v: v,
			})
		}
		return
	}))
	t.Run("key+%n", genTest(func(n int, key, flag string) (ret []pair) {
		for i := 0; i < n; i++ {
			v := crc32.ChecksumIEEE([]byte(fmt.Sprintf("%s#%d", key, i)))
			ret = append(ret, pair{
				f: flag,
				v: v,
			})
		}
		return
	}))
	t.Run("%0n+key", genTest(func(n int, key, flag string) (ret []pair) {
		for i := 0; i < n; i++ {
			v := crc32.ChecksumIEEE([]byte(fmt.Sprintf("%03d#%s", i, key)))
			ret = append(ret, pair{
				f: flag,
				v: v,
			})
		}
		return
	}))
	t.Run("key+%0n", genTest(func(n int, key, flag string) (ret []pair) {
		for i := 0; i < n; i++ {
			v := crc32.ChecksumIEEE([]byte(fmt.Sprintf("%s#%03d", key, i)))
			ret = append(ret, pair{
				f: flag,
				v: v,
			})
		}
		return
	}))

	t.Run("key+%0n", genTest(func(n int, key, flag string) (ret []pair) {
		for i := 0; i < n; i++ {
			v := crc32.ChecksumIEEE([]byte(fmt.Sprint(i) + key))
			// v := crc32.ChecksumIEEE([]byte(fmt.Sprintf("%03d#%s", i, key)))
			//v := crc32.ChecksumIEEE([]byte(fmt.Sprintf("%03d#%s", i, kf)))
			//v := crc32.ChecksumIEEE([]byte(fmt.Sprintf("%s#%d", key, i)))
			ret = append(ret, pair{
				f: flag,
				v: v,
			})
		}
		return
	}))
}

func TestCH(t *testing.T) {
	list := make([]pair, 0, 512)

	// make data
	list = append(list, generateCrc(32, uuid.New().String(), "a")...)
	list = append(list, generateCrc(32, uuid.New().String(), "b")...)
	list = append(list, generateCrc(32, uuid.New().String(), "c")...)
	//list = append(list, generateCrc(t, 32, uuid.New().String(), "d")...)
	sort.Slice(list, func(i, j int) bool {
		return list[i].v < list[j].v
	})

	f := "-"
	count := 0
	sum := uint32(0)
	last := 0
	// print begin and end number
	t.Log(list[0].v, list[len(list)-1].v)
	minus := float64(list[len(list)-1].v) - float64(list[0].v)
	// list = append(list,
	// 	pair{
	// 		f: "-",
	// 		v: 0,
	// 	},
	// 	pair{
	// 		f: "-",
	// 		v: list[len(list)-1].v,
	// 	},
	// )
	// sort.Slice(list, func(i, j int) bool {
	// 	return list[i].v < list[j].v
	// })
	mk := make(map[string]uint32)
	for k, v := range list {
		if f != v.f || k == len(list)-1 {
			t.Logf("flag: %1s %1d %10d %10d\n", f, count, v.v, v.v-uint32(last))
			f = v.f
			count = 1
			mk[f] += v.v - uint32(last)
			sum += v.v - uint32(last)
			// sum = 0
			last = int(v.v)
			continue
		}
		count++
		// sum += int(v.v) - last
		// last = int(v.v)
	}
	//mk[f] += v.v - uint32(last)
	t.Logf("flag: %1s %1d %10d %10d\n", f, count, last, 0)
	// t.Log(list)
	t.Log(sum, uint32(minus))
	rand.Seed(time.Now().UnixNano())

	mk = make(map[string]uint32)
	for k := 0; k < 1000000; k++ {
		n := crc32.ChecksumIEEE([]byte(fmt.Sprintf("%d", 100000+rand.Int()%1000000)))
		id := sort.Search(len(list), func(i int) bool {
			return n < list[i].v
		})
		if id >= len(list) {
			id = 0
		}
		mk[(list[id].f)] += 1
	}
	t.Log(mk, math.MaxUint32, math.MaxUint32/1024, math.MaxUint32>>7, math.MaxUint32-list[len(list)-1].v)

	limit := math.MaxUint32 / 1024

	mk = make(map[string]uint32)
	for _, v := range list {
		t.Log("pf", v.f, v.v/uint32(limit))
	}
	final := make([]string, 1024)
	begin := int(list[0].v / uint32(limit))
	for k := 0; k < int(begin); k++ {
		final[k] = list[len(list)-1].f
		mk[list[len(list)-1].f]++
	}
	for k := 0; k < len(list)-1; k++ {
		up := list[k+1].v / uint32(limit)
		for i := begin; i < int(up); i++ {
			final[i] = list[k].f
			mk[list[k].f]++
		}
		begin = int(up)
	}
	for k := list[len(list)-1].v / uint32(limit); k < 1024; k++ {
		final[k] = list[len(list)-1].f
		mk[list[len(list)-1].f]++
	}

	t.Log("all:", final, mk)
}

func checkMod(tb testing.TB, n int64) bool {
	v1 := n % 16384
	v2 := n & 0x3fff
	return v1 == v2
}

func TestMod(t *testing.T) {
	assert.True(t, checkMod(t, rand.Int63()), "rand int64 calc")
}

func BenchmarkMod(b *testing.B) {
	b.Run("compare", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if !checkMod(b, int64(i)) {
				b.Fail()
			}
		}
	})
	b.Run("mod", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = i % 16384
		}
	})
	b.Run("bit", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = i & 0x3fff
		}
	})
}

func BenchmarkAN(b *testing.B) {
	for i := 0; i < b.N; i++ {
		j := i
		if i != (j ^ 0) {
			b.Fail()
		}
	}
}
