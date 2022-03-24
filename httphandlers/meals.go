package httphandlers

import (
	"encoding/json"
	"fmt"
	"github.com/MeetPlan/MeetPlanBackend/sql"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

type MealJSON struct {
	sql.Meal
	HasOrdered bool
	MealOrders []UserJSON
}

type MealDate struct {
	Date  string
	Meals []MealJSON
}

func (server *httpImpl) GetMeals(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	userId, err := strconv.Atoi(fmt.Sprint(jwt["user_id"]))
	if err != nil {
		WriteBadRequest(w)
		return
	}
	meals, err := server.db.GetMeals()
	if err != nil {
		WriteJSON(w, Response{Success: false, Error: err.Error()}, http.StatusInternalServerError)
		return
	}
	var mealJson = make([]MealDate, 0)
	for i := 0; i < len(meals); i++ {
		meal := meals[i]
		var orders []int
		err := json.Unmarshal([]byte(meal.Orders), &orders)
		if err != nil {
			WriteJSON(w, Response{Success: false, Error: err.Error()}, http.StatusInternalServerError)
			return
		}
		var ordered = contains(orders, userId)
		var hasAppended = false
		var mealOrders = make([]UserJSON, 0)
		if jwt["role"] == "admin" {
			for n := 0; n < len(orders); n++ {
				user, err := server.db.GetUser(orders[n])
				if err != nil {
					return
				}
				mealOrders = append(mealOrders, UserJSON{
					Name:  user.Name,
					ID:    user.ID,
					Email: user.Email,
					Role:  user.Role,
				})
			}
		}
		for n := 0; n < len(mealJson); n++ {
			if mealJson[n].Date == meal.Date {
				mealJson[n].Meals = append(mealJson[n].Meals, MealJSON{
					Meal:       meal,
					HasOrdered: ordered,
					MealOrders: mealOrders,
				})
				hasAppended = true
				break
			}
		}
		if !hasAppended {
			meals := make([]MealJSON, 0)
			meals = append(meals, MealJSON{
				Meal:       meal,
				HasOrdered: ordered,
				MealOrders: mealOrders,
			})
			mealJson = append(mealJson, MealDate{
				Date:  meal.Date,
				Meals: meals,
			})
		}
	}
	for i, j := 0, len(mealJson)-1; i < j; i, j = i+1, j-1 {
		mealJson[i], mealJson[j] = mealJson[j], mealJson[i]
	}
	WriteJSON(w, Response{Success: true, Data: mealJson}, http.StatusOK)
}

func (server *httpImpl) NewMeal(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if jwt["role"] != "admin" {
		WriteForbiddenJWT(w)
		return
	}
	price, err := strconv.ParseFloat(r.FormValue("price"), 32)
	if err != nil {
		return
	}
	isLimited, err := strconv.ParseBool(r.FormValue("isLimited"))
	if err != nil {
		return
	}
	isVegan, err := strconv.ParseBool(r.FormValue("isVegan"))
	if err != nil {
		return
	}
	isVegetarian, err := strconv.ParseBool(r.FormValue("isVegetarian"))
	if err != nil {
		return
	}
	isLactoseFree, err := strconv.ParseBool(r.FormValue("isLactoseFree"))
	if err != nil {
		return
	}
	orderLimit, err := strconv.Atoi(r.FormValue("limit"))
	if err != nil {
		return
	}
	meal := sql.Meal{
		ID:            server.db.GetLastMealID(),
		Meals:         r.FormValue("description"),
		Date:          r.FormValue("date"),
		MealTitle:     r.FormValue("title"),
		Price:         float32(price),
		Orders:        "[]",
		IsLimited:     isLimited,
		OrderLimit:    orderLimit,
		IsVegan:       isVegan,
		IsVegetarian:  isVegetarian,
		IsLactoseFree: isLactoseFree,
		BlockOrders:   false,
	}
	err = server.db.InsertMeal(meal)
	if err != nil {
		return
	}
	WriteJSON(w, Response{Success: true, Data: "OK"}, http.StatusCreated)
}

func (server *httpImpl) NewOrder(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	userId, err := strconv.Atoi(fmt.Sprint(jwt["user_id"]))
	if err != nil {
		WriteBadRequest(w)
		return
	}
	mealId, err := strconv.Atoi(mux.Vars(r)["meal_id"])
	if err != nil {
		WriteBadRequest(w)
		return
	}
	meal, err := server.db.GetMeal(mealId)
	if err != nil {
		return
	}
	if meal.BlockOrders {
		WriteJSON(w, Response{Success: false, Data: "Orders are blocked."}, http.StatusConflict)
		return
	}
	var orders []int
	err = json.Unmarshal([]byte(meal.Orders), &orders)
	if err != nil {
		return
	}
	if contains(orders, userId) {
		WriteJSON(w, Response{Success: false, Data: "You cannot order same meal twice."}, http.StatusConflict)
		return
	}
	if meal.IsLimited && len(orders) >= meal.OrderLimit {
		WriteJSON(w, Response{Success: false, Data: "Orders are closed - maximum orders reached."}, http.StatusConflict)
		return
	}
	orders = append(orders, userId)
	marshal, err := json.Marshal(orders)
	if err != nil {
		return
	}
	meal.Orders = string(marshal)
	err = server.db.UpdateMeal(meal)
	if err != nil {
		return
	}
	WriteJSON(w, Response{Success: true, Data: "OK"}, http.StatusCreated)
}

func (server *httpImpl) EditMeal(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if jwt["role"] != "admin" {
		WriteForbiddenJWT(w)
		return
	}
	//userId, err := strconv.Atoi(fmt.Sprint(jwt["user_id"]))
	//if err != nil {
	//	WriteBadRequest(w)
	//	return
	//}
	mealId, err := strconv.Atoi(mux.Vars(r)["meal_id"])
	if err != nil {
		WriteBadRequest(w)
		return
	}
	meal, err := server.db.GetMeal(mealId)
	if err != nil {
		return
	}
	price, err := strconv.ParseFloat(r.FormValue("price"), 32)
	if err != nil {
		return
	}
	isLimited, err := strconv.ParseBool(r.FormValue("isLimited"))
	if err != nil {
		return
	}
	isVegan, err := strconv.ParseBool(r.FormValue("isVegan"))
	if err != nil {
		return
	}
	isVegetarian, err := strconv.ParseBool(r.FormValue("isVegetarian"))
	if err != nil {
		return
	}
	isLactoseFree, err := strconv.ParseBool(r.FormValue("isLactoseFree"))
	if err != nil {
		return
	}
	orderLimit, err := strconv.Atoi(r.FormValue("limit"))
	if err != nil {
		return
	}
	meal.Price = float32(price)
	meal.IsLimited = isLimited
	meal.IsVegan = isVegan
	meal.IsVegetarian = isVegetarian
	meal.IsLactoseFree = isLactoseFree
	meal.OrderLimit = orderLimit
	meal.Meals = r.FormValue("description")
	meal.Date = r.FormValue("date")
	meal.MealTitle = r.FormValue("title")
	err = server.db.UpdateMeal(meal)
	if err != nil {
		return
	}
	WriteJSON(w, Response{Success: true, Data: "OK"}, http.StatusCreated)
}

func (server *httpImpl) DeleteMeal(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if jwt["role"] != "admin" {
		WriteForbiddenJWT(w)
		return
	}
	//userId, err := strconv.Atoi(fmt.Sprint(jwt["user_id"]))
	//if err != nil {
	//	WriteBadRequest(w)
	//	return
	//}
	mealId, err := strconv.Atoi(mux.Vars(r)["meal_id"])
	if err != nil {
		WriteBadRequest(w)
		return
	}
	err = server.db.DeleteMeal(mealId)
	if err != nil {
		return
	}
	WriteJSON(w, Response{Success: true, Data: "OK"}, http.StatusCreated)
}

func (server *httpImpl) BlockUnblockOrder(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if jwt["role"] != "admin" {
		WriteForbiddenJWT(w)
		return
	}
	//userId, err := strconv.Atoi(fmt.Sprint(jwt["user_id"]))
	//if err != nil {
	//	WriteBadRequest(w)
	//	return
	//}
	mealId, err := strconv.Atoi(mux.Vars(r)["meal_id"])
	if err != nil {
		WriteBadRequest(w)
		return
	}
	meal, err := server.db.GetMeal(mealId)
	if err != nil {
		return
	}
	meal.BlockOrders = !meal.BlockOrders
	err = server.db.UpdateMeal(meal)
	if err != nil {
		return
	}
	WriteJSON(w, Response{Success: true, Data: "OK"}, http.StatusCreated)
}
