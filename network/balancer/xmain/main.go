package main

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"math"
	"math/rand"
	"sort"
	"time"

	"github.com/google/uuid"
)

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

var genKey = func(i int) string {
	return uuid.NewString()
}

var makePairList = func(keyCount, itemCount int) []pair {
	list := make([]pair, 0, keyCount*itemCount)
	for i := 0; i < keyCount; i++ {
		list = append(list, generateCrc(itemCount, genKey(i), fmt.Sprintf("%c", 'A'+i))...)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].v < list[j].v
	})
	return list
}

var appendIndex int
var genAppendKey = func() (key, flag string) {
	appendIndex++
	return uuid.NewString(), "-" //string('A' + appendIndex - 1)
}

var appendPairList = func(list []pair, itemCount int) []pair {
	k, f := genAppendKey()
	list = append(list, generateCrc(itemCount, k, f)...)
	sort.Slice(list, func(i, j int) bool {
		return list[i].v < list[j].v
	})
	return list
}

var makeRingArray = func(list []pair, arraySize int) (ring []string, collection map[string]int) {
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
	for k := list[len(list)-1].v / uint32(limit); k < uint32(arraySize); k++ {
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

var (
	// 分片数量
	slice int = 1024
	// 运行次数
	N int = 1e6
	// 虚拟节点数量
	virtual = 32
)
var (
	p3 = float64(slice) / 3.0
	p4 = float64(slice) / 4.0
)

var calcPercent = func() (min3, max3, min4, max4, change float64) {
	list1 := makePairList(3, virtual)
	_, c1024 := makeRingArray(list1, slice)
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

	list2 := appendPairList(list1, virtual)
	_, c1024c := makeRingArray(list2, slice)
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
	change = float64(c1024c["-"]) / float64(slice) * 100
	if N < 2 {
		fmt.Println(c1024)
		fmt.Println(c1024c)
	}
	return
}

var hashNum = crc32.ChecksumIEEE

var runTest = func(flag string) {
	fmt.Println(flag)
	min3, max3, min4, max4, change := math.MaxFloat64, 0.0, math.MaxFloat64, 0.0, 0.0
	cmin := math.MaxFloat64
	cmax := 0.0
	start := time.Now()
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
		if a5 > cmax {
			cmax = a5
		}
		if a5 < cmin {
			cmin = a5
		}
	}
	fmt.Printf(",%8d,%2.2f%%,%2.2f%%,%2.2f%%, %2.2f%%, %2.3f%%,%2.3f%%,%2.3f%%,%v\n",
		int64(N),
		min3, max3, min4, max4,
		change/(float64(N)),
		cmin, cmax,
		time.Now().Sub(start),
	)
}

var makeKey = func(key string, n int) []byte {
	return []byte(fmt.Sprint(n) + "#" + key)
}

func run4test() {
	// n#key
	makeKey = func(key string, n int) []byte {
		return []byte(fmt.Sprint(n) + "#" + key)
	}
	runTest("n#key")
	// key#n
	makeKey = func(key string, n int) []byte {
		return []byte(fmt.Sprintf("%s#%d", key, n))
	}
	runTest("key#n")
	// %03d#key
	makeKey = func(key string, n int) []byte {
		return []byte(fmt.Sprintf("%03d#%s", n, key))
	}
	runTest("%03d#key")
	// key#%03d
	makeKey = func(key string, n int) []byte {
		return []byte(fmt.Sprintf("%s#%03d", key, n))
	}
	runTest("key#%03d")
}

func main() {

	// 打印title
	fmt.Println("tip,times,3min,3max,4min,4max,change,cmin,cmax,dur")

	// 随机数
	rand.Seed(time.Now().Unix())

	////////////////////////////////////////////////////////////
	// uuid + 1024
	N = 1000
	fmt.Printf("key=%s slice=%d visual=%d p3=%.2f p4=%.2f\n",
		"uuid",
		slice,virtual,p3,p4,
	)
	run4test()

	////////////////////////////////////////////////////////////
	// uuid + 1<<32
	slice = 1<<32

	p3 = float64(slice) / 3.0
	p4 = float64(slice) / 4.0

	makeRingArray = func(list []pair, arraySize int) (ring []string, collection map[string]int) {
		collection = make(map[string]int)
		begin := int(list[0].v)
		collection[list[len(list)-1].f] += begin
		for k := 0; k < len(list)-1; k++ {
			up := int(list[k+1].v)
			collection[list[k].f] += up - begin
			begin = int(up)
		}
		collection[list[len(list)-1].f] += 1<<32 - int(list[len(list)-1].v)

		return
	}
fmt.Printf("key=%s slice=%d visual=%d p3=%.2f p4=%.2f\n",
		"uuid",
		slice,virtual,p3,p4,
	)
	run4test()

	////////////////////////////////////////////////////////////
	// ip + 1024
	

	genKey = func(i int) string {
		return fmt.Sprintf("10.0.1.%d", i)
	}
	genAppendKey = func() (key string, flag string) {
		return fmt.Sprintf("10.0.2.%d", rand.Int31()), "-"
	}
	// 分片数量
	slice = 1024 //1 << 32 //1024
	// 运行次数
	N = 1e6
	// 虚拟节点
	virtual = 32

	p3 = float64(slice) / 3.0
	p4 = float64(slice) / 4.0

	// fmt.Printf("p3:%d p4:%d\n", int(p3), int(p4))

	// makeRingArray = func(list []pair, arraySize int) (ring []string, collection map[string]int) {
	// 	collection = make(map[string]int)
	// 	begin := int(list[0].v)
	// 	collection[list[len(list)-1].f] += begin
	// 	for k := 0; k < len(list)-1; k++ {
	// 		up := int(list[k+1].v)
	// 		collection[list[k].f] += up - begin
	// 		begin = int(up)
	// 	}
	// 	collection[list[len(list)-1].f] += 1<<32 - int(list[len(list)-1].v)

	// 	return
	// }

	hashNum = func(data []byte) uint32 {
		tmp := md5.Sum(data)
		return binary.LittleEndian.Uint32(tmp[:])
	}

	generateCrc = func(n int, key, flag string) (ret []pair) {
		for i := 0; i < n; i++ {
			v := hashNum(makeKey(key, i))
			ret = append(ret, pair{
				f: flag,
				v: v,
			})
		}
		return
	}

	generateCrc = func(n int, key, flag string) (ret []pair) {
		for i := 0; i < n/4; i++ {
			bn := md5.Sum(makeKey(key, i))
			ret = append(ret, pair{
				f: flag,
				v: binary.LittleEndian.Uint32(bn[:]),
			})
			ret = append(ret, pair{
				f: flag,
				v: binary.LittleEndian.Uint32(bn[1*4:]),
			})
			ret = append(ret, pair{
				f: flag,
				v: binary.LittleEndian.Uint32(bn[2*4:]),
			})
			ret = append(ret, pair{
				f: flag,
				v: binary.LittleEndian.Uint32(bn[3*4:]),
			})
		}
		return
	}

}

// %n+key
//     %n+key times: 1000001 3(4.10, 239.36) 4(2.73, 267.58) change:25.011
// key+%n
//     key+%n times: 1000001 3(8.79, 222.36) 4(5.86, 248.44) change:24.995
// %0n+key
//    %0n+key times: 1000001 3(4.10, 220.90) 4(0.78, 283.20) change:25.001
// key+%0n
//    key+%0n times: 1000001 3(9.38, 249.02) 4(4.69, 276.17) change:25.005

// 1247508173/1431655765
