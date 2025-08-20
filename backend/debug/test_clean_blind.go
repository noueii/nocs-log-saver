package main
import (
	"fmt"
	cs2log "github.com/janstuemmel/cs2-log"
)
func main() {
	test1 := `L 08/19/2025 - 19:02:50: "SHESKY<7><[U:1:215888626]><TERRORIST>" blinded for 5.09 by "SHESKY<7><[U:1:215888626]><TERRORIST>" from flashbang entindex 225`
	test2 := `L 08/19/2025 - 19:02:50: "SHESKY<7><[U:1:215888626]><TERRORIST>" blinded for 5.09 by "Attacker<8><[U:1:123456789]><TERRORIST>" from flashbang entindex 225`
	
	parsed1, err1 := cs2log.Parse(test1)
	fmt.Printf("Test 1 (clean): Type=%T, Err=%v\n", parsed1, err1)
	
	parsed2, err2 := cs2log.Parse(test2)
	fmt.Printf("Test 2 (clean): Type=%T, Err=%v\n", parsed2, err2)
}
