package tool

import (
	"encoding/hex"
	"fmt"
	"testing"
)

func TestName(t *testing.T) {
	content := "200fab40f2410122500b95d9b4cb6a0097affbb04b872233fb4ca5a7972bdddb10ac0063036f7264010118746578742f706c61696e3b636861727365743d7574662d38004d4a017b22636861696e466c6167223a22627463222c2264617461223a226b796c652e6f7264222c226461746154797065223a22746578742f706c61696e222c22656e636f64696e67223a227574662d38222c22656e6372797074223a2230222c226d6574614944466c6167223a226d6574616964222c226e6f64654e616d65223a226e616d65222c2270223a226d6574616964222c22706172656e7454784964223a226274633a37633461336266346630656238336230633337336362306132313937333032336533623531343036303034366263333965363764383765613865356531373136222c227075626c69634b6579223a22303239653964383630333361346130646536303834396263386463636265346333643165386531663566653832306433666465663566373562396366366138373836222c2276657273696f6e223a22312e302e30227d68"

	b, _ := hex.DecodeString(content)
	fmt.Println(string(b))
}