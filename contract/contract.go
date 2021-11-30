package develogcontract

import (
	"github.com/nspcc-dev/neo-go/pkg/interop/native/oracle"
	"github.com/nspcc-dev/neo-go/pkg/interop/native/std"
	"github.com/nspcc-dev/neo-go/pkg/interop/runtime"
	"github.com/nspcc-dev/neo-go/pkg/interop/storage"
)

type Player struct {
	balance int
	gear    []int
}

func player(playerID string) Player {
	ctx := storage.GetContext()
	data := storage.Get(ctx, playerID)
	if data == nil {
		panic("player is not exists")
	}

	return std.Deserialize(data.([]byte)).(Player)
}

func _deploy(data interface{}, isUpdate bool) {
	if isUpdate {
		return
	}

	marie := Player{
		balance: 3000,
		gear:    []int{},
	}

	ctx := storage.GetContext()
	storage.Put(ctx, "Marie", std.Serialize(marie))

	containerID := "FjD1bdfxLLuHz4xYGzRmbWqdsPagSifX9UyQFA427D8k"
	objectID := "65nH4W3Anj17xcDj97LkHW7rW3GiUeRcP5tgnj7s5qr3"
	requestURI := "neofs:" + containerID + "/" + objectID

	storage.Put(ctx, "db", requestURI)
}

func Balance(playerID string) int {
	p := player(playerID)
	return p.balance
}

func Items(playerID string) []int {
	p := player(playerID)
	return p.gear
}

func BuyItem(playerID string, itemID int) {
	p := player(playerID)
	for i := range p.gear {
		if itemID == p.gear[i] {
			panic("item already purchased")
		}
	}

	ctx := storage.GetContext()
	uri := storage.Get(ctx, "db").(string)

	filter := []byte("$.store.item[" + std.Itoa10(itemID) + "]")

	oracle.Request(uri, filter, "buyItemCB", playerID, 2*oracle.MinimumResponseGas)
}

func BuyItemCB(url string, userData interface{}, code int, result []byte) {
	// This function shouldn't be called directly, we only expect oracle native
	// contract to be calling it.
	if string(runtime.GetCallingScriptHash()) != oracle.Hash {
		panic("not called from oracle contract")
	}
	if code != oracle.Success {
		panic("request failed for " + url + " with code " + std.Itoa(code, 10))
	}

	playerID := userData.(string)
	ln := len(result)
	data := std.JSONDeserialize(result[1 : ln-1]).(map[string]interface{})
	p := player(playerID)

	price := data["price"].(int)

	if p.balance < price {
		panic("not enough balance")
	}

	p.balance -= price
	p.gear = append(p.gear, data["id"].(int))

	ctx := storage.GetContext()
	storage.Put(ctx, playerID, std.Serialize(p))
}
