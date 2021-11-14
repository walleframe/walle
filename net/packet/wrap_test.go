package packet

import "testing"

func BenchmarkResetMessage(b *testing.B) {
	b.Run("Reset", func(b *testing.B) {
		pkg := &Packet{}
		for k := 0; k < b.N; k++ {
			pkg.Reset()
		}
	})
	b.Run("hand", func(b *testing.B) {
		v := &Packet{}
		for k := 0; k < b.N; k++ {
			v.Flag = 0
			v.Cmd = 0
			v.Sequence = 0
			v.Metadata = nil
			v.Body = nil
			v.Uri = ""
			v.ReservedRq = 0
		}
	})
}
