package dbhelper

import (
	"database/sql"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"test/database"
	"test/models"
	"test/utils"
	"time"
)

func IsUserExist(email string) (bool, error) {
	SQL := `SELECT id FROM users WHERE email = TRIM(LOWER($1)) AND archived_at is NULL`
	var id uuid.UUID
	err := database.Ecommerce.Get(&id, SQL, email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return false, err
	}
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return true, nil
}

func CreateUser(tx *sqlx.Tx, name, email, password, phone string) (uuid.UUID, error) {
	SQL := `INSERT INTO users(name,email,password,phone) VALUES ($1,$2,$3,$4) RETURNING id`
	var userid uuid.UUID
	err := tx.QueryRowx(SQL, name, email, password, phone).Scan(&userid)
	return userid, err
}

func CreateRole(tx *sqlx.Tx, userId uuid.UUID, role string) error {
	SQL := `INSERT INTO user_role(user_id,role) VALUES ($1,$2)`
	_, err := tx.Exec(SQL, userId, role)
	return err
}

func CreateCart(tx *sqlx.Tx, userId uuid.UUID) error {
	SQL := `INSERT INTO cart(user_id) VALUES ($1)`
	_, err := tx.Exec(SQL, userId)
	return err
}

func GetIdByPassword(email, password string) (uuid.UUID, error) {
	SQL := `SELECT id,password FROM users WHERE email = TRIM(LOWER($1)) AND archived_at is NULL`
	var userId uuid.UUID
	var pass string
	err := database.Ecommerce.QueryRowx(SQL, email).Scan(&userId, &pass)
	if err != nil {
		return userId, err
	}
	isMatch := utils.CheckPasswordHash(password, pass)
	if !isMatch {
		return uuid.Nil, nil
	}
	return userId, nil
}

func CheckRole(userId uuid.UUID) ([]string, error) {
	SQL := `SELECT ur.role FROM users u JOIN user_role ur ON ur.user_id = $1 WHERE  u.archived_at IS NULL `
	var role = make([]string, 0)
	err := database.Ecommerce.Select(&role, SQL, userId)
	return role, err
}

func CreateAddress(userId uuid.UUID, name, pinCode string, latitude, longitude float64) error {
	SQL := `INSERT INTO address(user_id,name,pin_code,latitude,longitude) VALUES ($1,$2,$3,$4,$5)`
	_, err := database.Ecommerce.Exec(SQL, userId, name, pinCode, latitude, longitude)
	return err
}

func GetAddress(userid uuid.UUID) ([]models.Address, error) {
	SQL := `SELECT id,name,pin_code,latitude,longitude FROM address WHERE user_id = $1 AND archived_at is NULL`
	var address = make([]models.Address, 0)
	err := database.Ecommerce.Select(&address, SQL, userid)
	return address, err
}

func DeleteAddress(userId, addressId uuid.UUID) (sql.Result, error) {
	SQL := `UPDATE address SET archived_at = $3 WHERE user_id = $1 AND id = $2 AND archived_at is NULL `
	result, err := database.Ecommerce.Exec(SQL, userId, addressId, time.Now())
	return result, err
}

func GetAllUser() ([]models.Users, error) {
	var users = make([]models.Users, 0)
	SQL := `SELECT id,name,email,password,phone,created_at from users where archived_at is NULL`
	err := database.Ecommerce.Select(&users, SQL)
	return users, err
}

func GetUser() ([]models.Users, error) {
	var users = make([]models.Users, 0)
	SQL := `SELECT u.id,u.name,u.email,u.password,u.phone,u.created_at FROM users u JOIN user_role ur ON u.id = ur.user_id WHERE ur.role != 'admin' AND u.archived_at is NULL`
	err := database.Ecommerce.Select(&users, SQL)
	return users, err
}

func DeleteUser(userId uuid.UUID) (sql.Result, error) {
	SQL := `UPDATE users SET archived_at = $1 where id = $2 AND archived_at is NULL`
	result, err := database.Ecommerce.Exec(SQL, time.Now(), userId)
	return result, err
}

func AddProduct(name string, cost int, itemType string, quantity int) error {
	SQL := `INSERT INTO product(name,cost,item_type,quantity) VALUES ($1,$2,$3,$4)`
	_, err := database.Ecommerce.Exec(SQL, name, cost, itemType, quantity)
	return err
}

func CreateUpload(pathname, filename, url string) (uuid.UUID, error) {
	SQL := `INSERT INTO upload(path,name,url) values ($1,$2,$3) RETURNING id`
	var uploadId uuid.UUID
	err := database.Ecommerce.QueryRowx(SQL, pathname, filename, url).Scan(&uploadId)
	return uploadId, err
}

func CreateUploadImage(uploadId, productId uuid.UUID) error {
	SQL := `INSERT INTO image(product_id,upload_id) VALUES ($1,$2)`
	_, err := database.Ecommerce.Exec(SQL, productId, uploadId)
	return err
}

func GetProduct() ([]models.Products, error) {
	SQL := `SELECT p.id,p.name,p.cost,p.item_type,p.quantity,array_agg(u.url) AS images FROM product p JOIN image i ON p.id = i.product_id JOIN upload u ON i.upload_id = u.id WHERE p.archived_at IS NULL group by p.id `
	var products = make([]models.Products, 0)
	err := database.Ecommerce.Select(&products, SQL)
	return products, err

}

func GetProductByType(itemType string) ([]models.Products, error) {
	SQL := `SELECT p.id,p.name,p.cost,p.item_type,p.quantity,array_agg(u.url) AS images FROM product p join image i on p.id = i.product_id JOIN upload u ON i.upload_id = u.id WHERE p.archived_at is NULL AND p.item_type = $1 group by p.id`
	var products = make([]models.Products, 0)
	err := database.Ecommerce.Select(&products, SQL, itemType)
	return products, err
}

func DeleteProductById(productId uuid.UUID) error {
	SQL := `UPDATE product SET archived_at = $2 WHERE id = $1 AND archived_at IS NULL`
	_, err := database.Ecommerce.Exec(SQL, productId, time.Now())
	return err
}

func CartIdFromUserId(userId uuid.UUID) (uuid.UUID, error) {
	SQL := `SELECT id FROM cart where user_id = $1`
	var cartId uuid.UUID
	err := database.Ecommerce.QueryRowx(SQL, userId).Scan(&cartId)
	return cartId, err
}

func FindProductFromCartItems(productId, cartId uuid.UUID) (bool, int, error) {
	SQL := `SELECT quantity FROM cart_items WHERE product_id = $1 AND cart_id = $2`
	var quantity int
	err := database.Ecommerce.QueryRowx(SQL, productId, cartId).Scan(&quantity)
	if err != nil && err != sql.ErrNoRows {
		return false, 0, err
	}
	if err == sql.ErrNoRows {
		return false, 0, nil
	}
	return true, quantity, nil
}

func InsertCartItem(cartId, productId uuid.UUID, quantity int) error {
	SQL := `INSERT INTO cart_items(cart_id,product_id,quantity) VALUES ($1,$2,$3)`
	_, err := database.Ecommerce.Exec(SQL, cartId, productId, quantity)
	return err
}

func IncrementCartItem(cartId, productId uuid.UUID, quantity int) error {
	SQL := `UPDATE cart_items SET quantity = $3 WHERE cart_id = $1 AND product_id = $2`
	_, err := database.Ecommerce.Exec(SQL, cartId, productId, quantity+1)
	return err
}

func DeleteCartItem(cartId, productId uuid.UUID) error {
	SQL := `DELETE FROM cart_items WHERE product_id = $1 AND cart_id = $2`
	_, err := database.Ecommerce.Exec(SQL, productId, cartId)
	return err
}

func DecrementCartItem(cartId, productId uuid.UUID, quantity int) error {
	SQL := `UPDATE cart_items SET quantity = $3 WHERE cart_id = $1 AND product_id = $2`
	_, err := database.Ecommerce.Exec(SQL, cartId, productId, quantity-1)
	return err
}

func ShowCartItems(cartId uuid.UUID) ([]models.CartItems, error) {
	var cartItems = make([]models.CartItems, 0)
	SQL := `SELECT product_id,quantity FROM cart_items WHERE cart_id = $1`
	err := database.Ecommerce.Select(&cartItems, SQL, cartId)
	return cartItems, err
}
