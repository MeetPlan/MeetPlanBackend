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
		for n := 0; n < len(mealJson); n++ {
			if mealJson[n].Date == meal.Date {
				mealJson[n].Meals = append(mealJson[n].Meals, MealJSON{
					Meal:       meal,
					HasOrdered: ordered,
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
			})
			mealJson = append(mealJson, MealDate{
				Date:  meal.Date,
				Meals: meals,
			})
		}
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
