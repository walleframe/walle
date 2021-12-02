package hash

import (
	"fmt"
	"hash/crc32"
	"math/rand"
	"sort"
	"testing"
	"time"

	"github.com/google/uuid"
)

type pair struct {
	v uint32
	f string
}

func TestCH(t *testing.T) {
	list := make([]pair, 0, 512)

	list = append(list, generateCrc(t, 16, uuid.New().String(), "a")...)
	list = append(list, generateCrc(t, 16, uuid.New().String(), "b")...)
	list = append(list, generateCrc(t, 16, uuid.New().String(), "c")...)
	sort.Slice(list, func(i, j int) bool {
		return list[i].v < list[j].v
	})
	f := "-"
	count := 0
	sum := uint32(0)
	last := 0
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
	t.Log(mk)
}

func generateCrc(t *testing.T, n int, key, flag string) (ret []pair) {
	for i := 0; i < n; i++ {
		// v := crc32.ChecksumIEEE([]byte(fmt.Sprint(i) + key))
		v := crc32.ChecksumIEEE([]byte(fmt.Sprintf("%3d#%s", i, key)))
		// v := crc32.ChecksumIEEE([]byte(fmt.Sprintf("%s#%d", key, i)))
		ret = append(ret, pair{
			f: flag,
			v: v,
		})
	}
	return
}
