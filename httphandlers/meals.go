package httphandlers

import (
	"encoding/json"
	"github.com/MeetPlan/MeetPlanBackend/helpers"
	"github.com/MeetPlan/MeetPlanBackend/sql"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

type MealJSON struct {
	sql.Meal
	HasOrdered     bool
	MealOrders     []UserJSON
	IsLimitReached bool
}

type MealDate struct {
	Date  string
	Meals []MealJSON
}

func (server *httpImpl) GetMeals(w http.ResponseWriter, r *http.Request) {
	if server.config.BlockMeals {
		WriteJSON(w, Response{Data: "Admin has disabled meals", Success: false}, http.StatusForbidden)
		return
	}
	user, err := server.db.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
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
		var ordered = helpers.Contains(orders, user.ID)
		var isLimitReached = meal.IsLimited && len(orders) >= meal.OrderLimit
		var hasAppended = false
		var mealOrders = make([]UserJSON, 0)
		if user.Role == ADMIN || user.Role == PRINCIPAL || user.Role == PRINCIPAL_ASSISTANT || user.Role == FOOD_ORGANIZER {
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
					Meal:           meal,
					HasOrdered:     ordered,
					MealOrders:     mealOrders,
					IsLimitReached: isLimitReached,
				})
				hasAppended = true
				break
			}
		}
		if !hasAppended {
			meals := make([]MealJSON, 0)
			meals = append(meals, MealJSON{
				Meal:           meal,
				HasOrdered:     ordered,
				MealOrders:     mealOrders,
				IsLimitReached: isLimitReached,
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
	if server.config.BlockMeals {
		WriteJSON(w, Response{Data: "Admin has disabled meals", Success: false}, http.StatusForbidden)
		return
	}
	user, err := server.db.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if user.Role == ADMIN || user.Role == PRINCIPAL_ASSISTANT || user.Role == PRINCIPAL || user.Role == FOOD_ORGANIZER {
		price, err := strconv.ParseFloat(r.FormValue("price"), 32)
		if err != nil {
			WriteJSON(w, Response{Success: false, Data: "Could not parse price", Error: r.FormValue("price")}, http.StatusBadRequest)
			return
		}
		isLimited, err := strconv.ParseBool(r.FormValue("isLimited"))
		if err != nil {
			WriteJSON(w, Response{Success: false, Data: "Could not parse isLimited", Error: r.FormValue("isLimited")}, http.StatusBadRequest)
			return
		}
		isVegan, err := strconv.ParseBool(r.FormValue("isVegan"))
		if err != nil {
			WriteJSON(w, Response{Success: false, Data: "Could not parse isVegan", Error: r.FormValue("isVegan")}, http.StatusBadRequest)
			return
		}
		isVegetarian, err := strconv.ParseBool(r.FormValue("isVegetarian"))
		if err != nil {
			WriteJSON(w, Response{Success: false, Data: "Could not parse isVegetarian", Error: r.FormValue("isVegetarian")}, http.StatusBadRequest)
			return
		}
		isLactoseFree, err := strconv.ParseBool(r.FormValue("isLactoseFree"))
		if err != nil {
			WriteJSON(w, Response{Success: false, Data: "Could not parse isLactoseFree", Error: r.FormValue("isLactoseFree")}, http.StatusBadRequest)
			return
		}
		orderLimit, err := strconv.Atoi(r.FormValue("limit"))
		if err != nil {
			WriteJSON(w, Response{Success: false, Data: "Could not parse limit", Error: r.FormValue("limit")}, http.StatusBadRequest)
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
			WriteJSON(w, Response{Success: false, Data: "Could not insert meal", Error: err.Error()}, http.StatusInternalServerError)
			return
		}
		WriteJSON(w, Response{Success: true, Data: "OK"}, http.StatusCreated)
	} else {
		WriteForbiddenJWT(w)
		return
	}
}

func (server *httpImpl) NewOrder(w http.ResponseWriter, r *http.Request) {
	if server.config.BlockMeals {
		WriteJSON(w, Response{Data: "Admin has disabled meals", Success: false}, http.StatusForbidden)
		return
	}
	user, err := server.db.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
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
	if helpers.Contains(orders, user.ID) {
		WriteJSON(w, Response{Success: false, Data: "You cannot order same meal twice."}, http.StatusConflict)
		return
	}
	if meal.IsLimited && len(orders) >= meal.OrderLimit {
		WriteJSON(w, Response{Success: false, Data: "Orders are closed - maximum orders reached."}, http.StatusConflict)
		return
	}
	orders = append(orders, user.ID)
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
	if server.config.BlockMeals {
		WriteJSON(w, Response{Data: "Admin has disabled meals", Success: false}, http.StatusForbidden)
		return
	}
	user, err := server.db.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if user.Role == ADMIN || user.Role == PRINCIPAL || user.Role == PRINCIPAL_ASSISTANT || user.Role == FOOD_ORGANIZER {
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
	} else {
		WriteForbiddenJWT(w)
		return
	}
}

func (server *httpImpl) DeleteMeal(w http.ResponseWriter, r *http.Request) {
	if server.config.BlockMeals {
		WriteJSON(w, Response{Data: "Admin has disabled meals", Success: false}, http.StatusForbidden)
		return
	}
	user, err := server.db.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if user.Role == ADMIN || user.Role == PRINCIPAL || user.Role == PRINCIPAL_ASSISTANT || user.Role == FOOD_ORGANIZER {
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
	} else {
		WriteForbiddenJWT(w)
		return
	}
}

func (server *httpImpl) BlockUnblockOrder(w http.ResponseWriter, r *http.Request) {
	if server.config.BlockMeals {
		WriteJSON(w, Response{Data: "Admin has disabled meals", Success: false}, http.StatusForbidden)
		return
	}
	user, err := server.db.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if user.Role == ADMIN || user.Role == PRINCIPAL_ASSISTANT || user.Role == PRINCIPAL || user.Role == FOOD_ORGANIZER {
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
	} else {
		WriteForbiddenJWT(w)
		return
	}
}

func (server *httpImpl) RemoveOrder(w http.ResponseWriter, r *http.Request) {
	if server.config.BlockMeals {
		WriteJSON(w, Response{Data: "Admin has disabled meals", Success: false}, http.StatusForbidden)
		return
	}
	user, err := server.db.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
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
	for i := 0; i < len(orders); i++ {
		if orders[i] == user.ID {
			orders = helpers.Remove(orders, i)
		}
	}
	marshal, err := json.Marshal(orders)
	if err != nil {
		return
	}
	meal.Orders = string(marshal)
	err = server.db.UpdateMeal(meal)
	if err != nil {
		return
	}
	WriteJSON(w, Response{Success: true, Data: "OK"}, http.StatusOK)
}

func (server *httpImpl) MealsBlocked(w http.ResponseWriter, r *http.Request) {
	_, err := server.db.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	WriteJSON(w, Response{Success: true, Data: server.config.BlockMeals}, http.StatusOK)
}
