package hrp

import (
	"math/rand"
	"strings"
	"time"
)

func contains(s []string, e string) bool {
	for _, a := range s {
		if strings.EqualFold(a, e) {
			return true
		}
	}
	return false
}

func shuffleCartesianProduct(slice []map[string]interface{}) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	for len(slice) > 0 {
		n := len(slice)
		randIndex := r.Intn(n)
		slice[n-1], slice[randIndex] = slice[randIndex], slice[n-1]
		slice = slice[:n-1]
	}
}

func genCartesianProduct(params [][]map[string]interface{}) []map[string]interface{} {
	var cartesianProduct []map[string]interface{}
	cartesianProduct = params[0]
	for i := 0; i < len(params)-1; i++ {
		var tempProduct []map[string]interface{}
		for _, param1 := range cartesianProduct {
			for _, param2 := range params[i+1] {
				tempProduct = append(tempProduct, mergeVariables(param1, param2))
			}
		}
		cartesianProduct = tempProduct
	}
	return cartesianProduct
}
