package sql

type Meal struct {
	ID            int
	Meals         string
	Date          string
	MealTitle     string `db:"meal_title"`
	Price         float32
	Orders        string
	IsLimited     bool `db:"is_limited"`
	OrderLimit    int  `db:"order_limit"`
	IsVegan       bool `db:"is_vegan"`
	IsVegetarian  bool `db:"is_vegetarian"`
	IsLactoseFree bool `db:"is_lactose_free"`
	BlockOrders   bool `db:"block_orders"`
}

func (db *sqlImpl) GetMeal(id int) (meal Meal, err error) {
	err = db.db.Get(&meal, "SELECT * FROM meals WHERE id=$1", id)
	return meal, err
}

func (db *sqlImpl) InsertMeal(meal Meal) (err error) {
	_, err = db.db.NamedExec(
		"INSERT INTO meals (id, meals, date, meal_title, price, is_vegan, is_vegetarian, is_lactose_free, orders, order_limit, is_limited, block_orders) VALUES (:id, :meals, :date, :meal_title, :price, :is_vegan, :is_vegetarian, :is_lactose_free, :orders, :order_limit, :is_limited, :block_orders)",
		meal)
	return err
}

func (db *sqlImpl) UpdateMeal(meal Meal) error {
	_, err := db.db.NamedExec(
		"UPDATE meals SET meals=:meals, date=:date, meal_title=:meal_title, price=:price, is_vegan=:is_vegan, is_vegetarian=:is_vegetarian, is_lactose_free=:is_lactose_free, orders=:orders, order_limit=:order_limit, is_limited=:is_limited, block_orders=:block_orders WHERE id=:id",
		meal)
	return err
}

func (db *sqlImpl) GetLastMealID() (id int) {
	err := db.db.Get(&id, "SELECT id FROM meals WHERE id = (SELECT MAX(id) FROM meals)")
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return 0
		}
		db.logger.Info(err)
		return -1
	}
	return id + 1
}

func (db *sqlImpl) GetMeals() (meals []Meal, err error) {
	err = db.db.Select(&meals, "SELECT * FROM meals ORDER BY id ASC")
	return meals, err
}

func (db *sqlImpl) DeleteMeal(ID int) error {
	_, err := db.db.Exec("DELETE FROM meals WHERE id=$1", ID)
	return err
}
