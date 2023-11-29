package process_test

import (
	"errors"
	"strconv"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/walleframe/walle/process"
	"github.com/walleframe/walle/testpkg"
)

func TestPacketDispatcherChain(t *testing.T) {
	tf := func(t *testing.T, num int, errIndex int) {
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		fc := testpkg.NewMockFuncCall(ctl)
		errFailed := errors.New("dispatch failed")
		lists := make([]process.DataDispatcherFilter, 0, num)
		for k := 0; k < num; k++ {
			index := k
			runFlag := false
			if errIndex >= 0 {
				if index < errIndex {
					runFlag = true
				}
			} else {
				runFlag = true
			}
			if runFlag {
				fc.EXPECT().Call(gomock.Any()).Do(func(data []byte) {
					v, err := strconv.Atoi(string(data))
					assert.Nil(t, err, "index %d(%s) convert error need nil.%v", index, string(data), err)
					assert.Equal(t, index-1, v, "index %d(%s) recv data invalid", index, string(data))
					t.Log("index", index, string(data))
				})
			}

			lists = append(lists, func(data []byte, next process.DataDispatcherFunc) (err error) {
				if errIndex >= 0 {
					if index == errIndex {
						return errFailed
					} else if index > errIndex {
						t.Error("index func must not run", index, string(data))
						return
					}
				}
				fc.Call(data)
				return next([]byte(strconv.Itoa(index)))
			})
		}

		f := process.DataDispatcherChain(lists...)

		if errIndex < 0 {
			fc.EXPECT().Call(gomock.Any()).Do(func(data []byte) {
				v, err := strconv.Atoi(string(data))
				assert.Nil(t, err, "final convert error need nil")
				assert.Equal(t, len(lists)-1, v, "final recv data invalid")
				t.Log("final", string(data))
			})
		}

		ret := f([]byte("-1"), func(data []byte) error {
			fc.Call(data)
			return nil
		})
		if errIndex < 0 {
			assert.Nil(t, ret, "error result need nil")
		} else {
			assert.Equal(t, ret, errFailed, "error result need equal.")
		}
	}

	datas := []struct {
		name     string
		num      int
		errIndex int
	}{
		{"zero", 0, -1},
		{"one", 1, -1},
		{"two", 2, -1},
		{"two", 5, -1},
		{"err1/1", 1, 0},
		{"err1/5", 5, 0},
		{"err2/3", 3, 1},
		{"err3/5", 5, 2},
	}
	for _, item := range datas {
		t.Run(item.name, func(t *testing.T) {
			tf(t, item.num, item.errIndex)
		})
	}

}
