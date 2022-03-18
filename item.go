package shopee

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	jsoniter "github.com/json-iterator/go"
)

var ErrUrlNotMatch = errors.New("url doesn't match")

var (
	// link dari apk android
	prodLinkAppRe = regexp.MustCompile(`^https?://(?:mall\.)?shopee\.[\w.]+/product/(\d+)/(\d+)`)
	// link dari web
	prodLinkWebRe = regexp.MustCompile(`^https?://(?:mall\.)?shopee\.[\w.]+/.+\.(\d+)\.(\d+)`)
)

// parse product url
func ParseProdURL(urlstr string) (shopid, itemid int64, err error) {
	for _, re := range [...]*regexp.Regexp{prodLinkAppRe, prodLinkWebRe} {
		if match := re.FindStringSubmatch(urlstr); len(match) > 0 {
			shopid, _ = strconv.ParseInt(match[1], 10, 64)
			itemid, _ = strconv.ParseInt(match[2], 10, 64)
			return
		}
	}

	return 0, 0, ErrUrlNotMatch
}

// shopee product item
type Item struct {
	json        jsoniter.Any
	modelsCache []Model
	tvarsCache  []TierVar
}

func (c Client) FetchItem(shopid, itemid int64) (Item, error) {
	resp, err := c.Client.R().
		SetQueryParams(map[string]string{
			"itemid": strconv.FormatInt(itemid, 10),
			"shopid": strconv.FormatInt(shopid, 10),
		}).
		Get("/api/v4/item/get")
	if err != nil {
		return Item{}, err
	}

	json := jsoniter.Get(resp.Body())
	if err := json.Get("error").GetInterface(); err != nil {
		return Item{}, fmt.Errorf("%v: %v", err, json.Get("error_msg").GetInterface())
	}
	return Item{json: json.Get("data")}, nil
}

func (c Client) FetchItemFromURL(urlstr string) (Item, error) {
	shopid, itemid, err := ParseProdURL(urlstr)
	if err != nil {
		return Item{}, err
	}
	return c.FetchItem(shopid, itemid)
}

func (i Item) ShopID() int64          { return i.json.Get("shopid").ToInt64() }
func (i Item) ItemID() int64          { return i.json.Get("itemid").ToInt64() }
func (i Item) PriceMin() int64        { return i.json.Get("price_min").ToInt64() }
func (i Item) PriceMax() int64        { return i.json.Get("price_max").ToInt64() }
func (i Item) Price() int64           { return i.json.Get("price").ToInt64() }
func (i Item) Stock() int             { return i.json.Get("stock").ToInt() }
func (i Item) Name() string           { return i.json.Get("name").ToString() }
func (i Item) IsFlashSale() bool      { return i.json.Get("flash_sale").GetInterface() != nil }
func (i Item) HasUpcomingFsale() bool { return i.json.Get("upcoming_flash_sale").GetInterface() != nil }

func (i Item) UpcomingFsaleStartTime() int64 {
	return i.json.Get("upcoming_flash_sale", "start_time").ToInt64()
}

func (i Item) CatIDs() []int64 {
	cats := i.json.Get("categories")
	out := make([]int64, cats.Size())
	for i := 0; i < cats.Size(); i++ {
		out[i] = cats.Get(i, "catid").ToInt64()
	}
	return out
}

func (i Item) CatNames() []string {
	cats := i.json.Get("categories")
	out := make([]string, cats.Size())
	for i := 0; i < cats.Size(); i++ {
		out[i] = cats.Get(i, "display_name").ToString()
	}
	return out
}

type TierVar struct {
	json      jsoniter.Any
	optsCache []string
}

func (i Item) TierVariations() []TierVar {
	if i.tvarsCache != nil {
		return i.tvarsCache
	}

	tvars := i.json.Get("tier_variations")
	out := make([]TierVar, tvars.Size())
	for i := 0; i < tvars.Size(); i++ {
		out[i] = TierVar{json: tvars.Get(i)}
	}
	i.tvarsCache = out
	return out
}

func (t TierVar) Name() string { return t.json.Get("name").ToString() }

func (t TierVar) Options() []string {
	if t.optsCache != nil {
		return t.optsCache
	}

	opts := t.json.Get("options")
	out := make([]string, opts.Size())
	for i := 0; i < opts.Size(); i++ {
		out[i] = opts.Get(i).ToString()
	}
	t.optsCache = out
	return out
}

type Model struct{ json jsoniter.Any }

func (i *Item) Models() []Model {
	if i.modelsCache != nil {
		return i.modelsCache
	}

	models := i.json.Get("models")
	out := make([]Model, models.Size())
	for i := 0; i < models.Size(); i++ {
		out[i] = Model{models.Get(i)}
	}
	i.modelsCache = out
	return out
}

func (m Model) ItemID() int64  { return m.json.Get("itemid").ToInt64() }
func (m Model) Name() string   { return m.json.Get("name").ToString() }
func (m Model) Stock() int     { return m.json.Get("stock").ToInt() }
func (m Model) ModelID() int64 { return m.json.Get("modelid").ToInt64() }
func (m Model) Price() int64   { return m.json.Get("price").ToInt64() }

type CheckoutableItem struct {
	Item
	chosenModel int
}

func ChooseModel(item Item, modelId int64) CheckoutableItem {
	var modelIndex int
	for i, m := range item.Models() {
		if m.ModelID() == modelId {
			modelIndex = i
			break
		}
	}
	return CheckoutableItem{item, modelIndex}
}

// indexes is tier variations index.
// indexes must be the same length as len(item.TierVariations())
func ChooseModelByTierVar(item Item, indexes []int) CheckoutableItem {
	tvars := item.TierVariations()
	if len(indexes) != len(tvars) {
		panic(fmt.Errorf("len of indexes: %d, len of tier vars: %d", len(indexes), len(tvars)))
	}
	tvarsName := make([]string, len(tvars))
	for i, opt_i := range indexes {
		tvarsName[i] = tvars[i].Options()[opt_i]
	}
	modelName := strings.Join(tvarsName, ",")

	var modelIndex int
	for i, m := range item.Models() {
		if m.Name() == modelName {
			modelIndex = i
			break
		}
	}
	return CheckoutableItem{item, modelIndex}
}

func (c CheckoutableItem) ChosenModel() Model { return c.Models()[c.chosenModel] }
