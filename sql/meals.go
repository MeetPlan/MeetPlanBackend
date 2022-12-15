package sql

type Meal struct {
	ID            string
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

	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
}

func (db *sqlImpl) GetMeal(id string) (meal Meal, err error) {
	err = db.db.Get(&meal, "SELECT * FROM meals WHERE id=$1", id)
	return meal, err
}

func (db *sqlImpl) InsertMeal(meal Meal) (err error) {
	_, err = db.db.NamedExec(
		"INSERT INTO meals (meals, date, meal_title, price, is_vegan, is_vegetarian, is_lactose_free, orders, order_limit, is_limited, block_orders) VALUES (:meals, :date, :meal_title, :price, :is_vegan, :is_vegetarian, :is_lactose_free, :orders, :order_limit, :is_limited, :block_orders)",
		meal)
	return err
}

func (db *sqlImpl) UpdateMeal(meal Meal) error {
	_, err := db.db.NamedExec(
		"UPDATE meals SET meals=:meals, date=:date, meal_title=:meal_title, price=:price, is_vegan=:is_vegan, is_vegetarian=:is_vegetarian, is_lactose_free=:is_lactose_free, orders=:orders, order_limit=:order_limit, is_limited=:is_limited, block_orders=:block_orders WHERE id=:id",
		meal)
	return err
}

func (db *sqlImpl) GetMeals() (meals []Meal, err error) {
	err = db.db.Select(&meals, "SELECT * FROM meals ORDER BY id ASC")
	return meals, err
}

func (db *sqlImpl) DeleteMeal(ID string) error {
	_, err := db.db.Exec("DELETE FROM meals WHERE id=$1", ID)
	return err
}
