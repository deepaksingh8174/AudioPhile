package handler

import (
	"database/sql"
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"io"
	"net/http"
	"os"
	"test/database"
	"test/database/dbhelper"
	"test/models"
	"test/utils"
	"time"
)

func RegisterUser(w http.ResponseWriter, r *http.Request) {
	var user models.User
	decodeErr := json.NewDecoder(r.Body).Decode(&user)
	if decodeErr != nil {
		utils.RespondJSON(w, http.StatusBadRequest, utils.Status{Message: "failed to parse request body"})
		return
	}
	exist, existErr := dbhelper.IsUserExist(user.Email)
	if existErr != nil {
		utils.RespondJSON(w, http.StatusBadRequest, utils.Status{Message: "failed to check user existence"})
		return
	}
	if exist {
		utils.RespondJSON(w, http.StatusConflict, utils.Status{Message: "user already exists"})
		return
	}

	hashedPassword, hashingErr := utils.HashPassword(user.Password)
	if hashingErr != nil {
		utils.RespondJSON(w, http.StatusInternalServerError, utils.Status{Message: "error occurred in password hashing"})
		return
	}

	txErr := database.Tx(func(tx *sqlx.Tx) error {
		userId, userErr := dbhelper.CreateUser(tx, user.Name, user.Email, hashedPassword, user.Phone)
		if userErr != nil {
			utils.RespondJSON(w, http.StatusInternalServerError, utils.Status{Message: "failed to create user"})
			return userErr
		}

		roleErr := dbhelper.CreateRole(tx, userId, models.UserRole)
		if roleErr != nil {
			utils.RespondJSON(w, http.StatusInternalServerError, utils.Status{Message: "failed to create user role"})
			return roleErr
		}

		cartErr := dbhelper.CreateCart(tx, userId)
		if cartErr != nil {
			utils.RespondJSON(w, http.StatusInternalServerError, utils.Status{Message: "failed to create cart"})
			return cartErr
		}
		return nil
	})
	if txErr != nil {
		utils.RespondJSON(w, http.StatusInternalServerError, utils.Status{Message: "transaction cannot be committed"})
		return
	}
	utils.RespondJSON(w, http.StatusCreated, utils.Status{Message: "user registered successfully"})
}

func Login(w http.ResponseWriter, r *http.Request) {
	var user models.Login
	decodeErr := json.NewDecoder(r.Body).Decode(&user)
	if decodeErr != nil {
		utils.RespondJSON(w, http.StatusBadRequest, utils.Status{Message: "failed to parse request body"})
		return
	}
	userId, userErr := dbhelper.GetIdByPassword(user.Email, user.Password)
	if userErr != nil {
		if userErr == sql.ErrNoRows {
			utils.RespondJSON(w, http.StatusBadRequest, utils.Status{Message: "wrong credentials entered"})
			return
		}
		utils.RespondJSON(w, http.StatusBadRequest, utils.Status{Message: "failed to process user database"})
		return
	}
	if userId == uuid.Nil {
		utils.RespondJSON(w, http.StatusBadRequest, utils.Status{Message: "wrong credentials entered"})
		return
	}
	claim := &models.Claims{
		UserId: userId,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	tokenString, tokenErr := token.SignedString([]byte(os.Getenv("secretKey")))
	if tokenErr != nil {
		utils.RespondJSON(w, http.StatusInternalServerError, utils.Status{Message: "error occurred in generation of token"})
		return
	}
	utils.RespondJSON(w, http.StatusOK, tokenString)
}

func CreateAddress(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("claims").(*models.Claims)
	userId := claims.UserId
	var address models.Address
	decodeErr := json.NewDecoder(r.Body).Decode(&address)
	if decodeErr != nil {
		utils.RespondJSON(w, http.StatusBadRequest, utils.Status{Message: "failed to parse request body"})
		return
	}
	addressErr := dbhelper.CreateAddress(userId, address.Name, address.PinCode, address.Latitude, address.Longitude)
	if addressErr != nil {
		utils.RespondJSON(w, http.StatusInternalServerError, utils.Status{Message: "failed to create user address"})
		return
	}
	utils.RespondJSON(w, http.StatusCreated, utils.Status{Message: "user address created"})
}

func ShowAddress(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("claims").(*models.Claims)
	userId := claims.UserId
	address, addressErr := dbhelper.GetAddress(userId)
	if addressErr != nil {
		utils.RespondJSON(w, http.StatusInternalServerError, utils.Status{Message: "Error Occurred during fetching of address table"})
		return
	}
	utils.RespondJSON(w, http.StatusOK, address)
}

func DeleteAddressById(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("claims").(*models.Claims)
	userId := claims.UserId
	id := r.URL.Query().Get("id")
	addressId, parseErr := uuid.Parse(id)
	if parseErr != nil {
		utils.RespondJSON(w, http.StatusInternalServerError, utils.Status{Message: "Cannot Parse string into uuid"})
		return
	}
	result, err := dbhelper.DeleteAddress(userId, addressId)
	if err != nil {
		utils.RespondJSON(w, http.StatusInternalServerError, utils.Status{Message: "failed to process query"})
		return
	}
	rowsAffected, rowsAffectedErr := result.RowsAffected()
	if rowsAffectedErr != nil {
		utils.RespondJSON(w, http.StatusInternalServerError, utils.Status{Message: "error occurred in rowsAffected function"})
		return
	}
	if rowsAffected == 0 {
		utils.RespondJSON(w, http.StatusBadRequest, utils.Status{Message: "enter a valid id"})
		return
	}
	utils.RespondJSON(w, http.StatusOK, utils.Status{Message: "address deleted successfully"})
}

func AddUser(w http.ResponseWriter, r *http.Request) {
	user := models.User{}
	decodeErr := json.NewDecoder(r.Body).Decode(&user)
	if decodeErr != nil {
		utils.RespondJSON(w, http.StatusBadRequest, utils.Status{Message: "failed to parse request body"})
		return
	}
	exist, existErr := dbhelper.IsUserExist(user.Email)
	if existErr != nil {
		utils.RespondJSON(w, http.StatusBadRequest, utils.Status{Message: "failed to check user existence"})
		return
	}
	if exist {
		utils.RespondJSON(w, http.StatusConflict, utils.Status{Message: "user already exists"})
		return
	}

	hashedPassword, hashingErr := utils.HashPassword(user.Password)
	if hashingErr != nil {
		utils.RespondJSON(w, http.StatusInternalServerError, utils.Status{Message: "error occurred in password hashing"})
		return
	}

	txErr := database.Tx(func(tx *sqlx.Tx) error {
		userId, userErr := dbhelper.CreateUser(tx, user.Name, user.Email, hashedPassword, user.Phone)
		if userErr != nil {
			utils.RespondJSON(w, http.StatusInternalServerError, utils.Status{Message: "failed to add user"})
			return userErr
		}

		roleErr := dbhelper.CreateRole(tx, userId, models.UserRole)
		if roleErr != nil {
			utils.RespondJSON(w, http.StatusInternalServerError, utils.Status{Message: "failed to create user role"})
			return roleErr
		}

		cartErr := dbhelper.CreateCart(tx, userId)
		if cartErr != nil {
			utils.RespondJSON(w, http.StatusInternalServerError, utils.Status{Message: "failed to create cart"})
			return cartErr
		}
		return nil
	})
	if txErr != nil {
		utils.RespondJSON(w, http.StatusInternalServerError, utils.Status{Message: "transaction cannot be committed"})
		return
	}
	utils.RespondJSON(w, http.StatusCreated, utils.Status{Message: "user added successfully"})
}

func ShowUser(w http.ResponseWriter, r *http.Request) {
	users, userErr := dbhelper.GetUser()
	if userErr != nil {
		utils.RespondJSON(w, http.StatusInternalServerError, utils.Status{Message: "error occurred during fetching of user table"})
		return
	}
	utils.RespondJSON(w, http.StatusOK, users)
}

func DeleteUserById(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	userId, parseErr := uuid.Parse(id)
	if parseErr != nil {
		utils.RespondJSON(w, http.StatusInternalServerError, utils.Status{Message: "cannot parse into uuid"})
		return
	}
	result, err := dbhelper.DeleteUser(userId)
	if err != nil {
		utils.RespondJSON(w, http.StatusInternalServerError, utils.Status{Message: "failed to parse query"})
	}
	rowAffected, rowAffectedErr := result.RowsAffected()
	if rowAffectedErr != nil {
		utils.RespondJSON(w, http.StatusInternalServerError, utils.Status{Message: "error occurred in rowsAffected function"})
		return
	}
	if rowAffected == 0 {
		utils.RespondJSON(w, http.StatusBadRequest, utils.Status{Message: "enter a valid id"})
		return
	}
	utils.RespondJSON(w, http.StatusOK, utils.Status{Message: "user deleted successfully"})
}

// check create product
func CreateProduct(w http.ResponseWriter, r *http.Request) {
	var product models.Product
	decodeErr := json.NewDecoder(r.Body).Decode(&product)
	if decodeErr != nil {
		utils.RespondJSON(w, http.StatusBadRequest, utils.Status{Message: "failed to parse request body"})
		return
	}
	err := dbhelper.AddProduct(product.Name, int(product.Cost), product.Type, product.Quantity)
	if err != nil {
		utils.RespondJSON(w, http.StatusInternalServerError, utils.Status{Message: "failed to parse query"})
		return
	}
	utils.RespondJSON(w, http.StatusCreated, utils.Status{Message: "product added successfully"})
}

func ShowProduct(w http.ResponseWriter, r *http.Request) {
	products, productErr := dbhelper.GetProduct()
	if productErr != nil {
		utils.RespondJSON(w, http.StatusInternalServerError, utils.Status{Message: "error occurred during fetching of product table"})
		return
	}
	utils.RespondJSON(w, http.StatusOK, products)
}

func ShowProductByType(w http.ResponseWriter, r *http.Request) {
	itemType := r.URL.Query().Get("item_type")
	if itemType != models.EarphoneType && itemType != models.HeadPhoneType && itemType != models.SpeakerType {
		utils.RespondJSON(w, http.StatusBadRequest, utils.Status{Message: "enter a valid type"})
		return
	}
	products, productErr := dbhelper.GetProductByType(itemType)
	if productErr != nil {
		utils.RespondJSON(w, http.StatusInternalServerError, utils.Status{Message: "error occurred during fetching of product table"})
		return
	}
	utils.RespondJSON(w, http.StatusOK, products)
}

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(30 << 20)
	if err != nil {
		utils.RespondJSON(w, http.StatusBadRequest, utils.Status{Message: "error occurred in ParseForm"})
		return
	}

	file, handler, errFile := r.FormFile("file")
	if errFile != nil {
		utils.RespondJSON(w, http.StatusBadRequest, utils.Status{Message: "error occurred in parsing the request"})
		return
	}
	defer file.Close()

	fileName := handler.Filename + time.Now().String()
	pathName := "./uploads/" + fileName
	f, openErr := os.OpenFile(pathName, os.O_WRONLY|os.O_CREATE, 0666)
	if openErr != nil {
		utils.RespondJSON(w, http.StatusInternalServerError, utils.Status{Message: "error occurred during opening of file"})
		return
	}
	defer f.Close()

	_, err = io.Copy(f, file)
	if err != nil {
		utils.RespondJSON(w, http.StatusInternalServerError, utils.Status{Message: "error occurred in copying of file"})
		return
	}

	uploadId, uploadErr := dbhelper.CreateUpload(pathName, fileName, "https://subham.jpg")
	if uploadErr != nil {
		utils.RespondJSON(w, http.StatusInternalServerError, utils.Status{Message: "error occurred in creation of upload image"})
		return
	}
	productIdString := r.URL.Query().Get("id")
	productId, parseErr := uuid.Parse(productIdString)
	if parseErr != nil {
		utils.RespondJSON(w, http.StatusInternalServerError, utils.Status{Message: "cannot parse string into uuid"})
		return
	}
	uploadImageErr := dbhelper.CreateUploadImage(uploadId, productId)
	if uploadImageErr != nil {
		utils.RespondJSON(w, http.StatusInternalServerError, utils.Status{Message: "query cannot be parsed"})
		return
	}
	utils.RespondJSON(w, http.StatusCreated, utils.Status{Message: "insertion in upload table and image table successfully."})
}

func UploadHandlerAWS(w http.ResponseWriter, r *http.Request) {
	parseErr := r.ParseMultipartForm(10 << 20)
	if parseErr != nil {
		utils.RespondJSON(w, http.StatusBadRequest, utils.Status{Message: "error occurred in ParseForm"})
		return
	}

	file, handler, errFile := r.FormFile("file")
	if errFile != nil {
		utils.RespondJSON(w, http.StatusBadRequest, utils.Status{Message: "error occurred in parsing the request"})
		return
	}
	defer file.Close()

	fileName := handler.Filename + time.Now().String()
	pathName := "./uploads/" + fileName

	AWSConfig := models.AWSConfig{
		AccessKeyID:     os.Getenv("accessKey"),
		AccessKeySecret: os.Getenv("secret_key"),
		Region:          os.Getenv("region"),
		BucketName:      os.Getenv("bucketName"),
	}

	sess := utils.CreateSession()

	uploader := s3manager.NewUploader(sess)
	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(AWSConfig.BucketName),
		Key:    aws.String(fileName),
		Body:   file,
	})

	if err != nil {
		utils.RespondJSON(w, http.StatusInternalServerError, utils.Status{Message: "error in creation of mew uploader session"})
		return
	}

	s3Client := utils.CreateS3Session(sess)

	req, _ := s3Client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: &AWSConfig.BucketName,
		Key:    &AWSConfig.AccessKeySecret,
	})
	urlStr, urlErr := req.Presign(20 * time.Minute)
	if urlErr != nil {
		utils.RespondJSON(w, http.StatusInternalServerError, utils.Status{Message: "Error occurred while generating url"})
		return
	}

	uploadId, uploadErr := dbhelper.CreateUpload(pathName, fileName, urlStr)
	if uploadErr != nil {
		utils.RespondJSON(w, http.StatusInternalServerError, utils.Status{Message: "error occurred in creation of upload image"})
		return
	}
	productIdString := r.URL.Query().Get("id")
	productId, parsingErr := uuid.Parse(productIdString)
	if parsingErr != nil {
		utils.RespondJSON(w, http.StatusInternalServerError, utils.Status{Message: "cannot parse string into uuid"})
		return
	}
	uploadImageErr := dbhelper.CreateUploadImage(uploadId, productId)
	if uploadImageErr != nil {
		utils.RespondJSON(w, http.StatusInternalServerError, utils.Status{Message: "query cannot be parsed"})
		return
	}
	utils.RespondJSON(w, http.StatusCreated, utils.Status{Message: "insertion in upload table and image table successfully."})
}

func DeleteProductById(w http.ResponseWriter, r *http.Request) {
	idString := r.URL.Query().Get("id")
	productId, productErr := uuid.Parse(idString)
	if productErr != nil {
		utils.RespondJSON(w, http.StatusInternalServerError, utils.Status{Message: "error occurred in parsing uuid"})
		return
	}
	err := dbhelper.DeleteProductById(productId)
	if err != nil {
		utils.RespondJSON(w, http.StatusInternalServerError, utils.Status{Message: "error occurred in processing query"})
		return
	}
	utils.RespondJSON(w, http.StatusOK, utils.Status{Message: "product deleted successfully"})
}

func AddCartItem(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("claims").(*models.Claims)
	userId := claims.UserId
	cartId, CartErr := dbhelper.CartIdFromUserId(userId)
	if CartErr != nil {
		utils.RespondJSON(w, http.StatusInternalServerError, utils.Status{Message: "error occurred in parsing table"})
		return
	}
	productIdString := r.URL.Query().Get("id")
	productId, parseErr := uuid.Parse(productIdString)
	if parseErr != nil {
		utils.RespondJSON(w, http.StatusInternalServerError, utils.Status{Message: "error occurred into parsing of uuid"})
		return
	}
	find, quantity, err := dbhelper.FindProductFromCartItems(productId, cartId)
	if err != nil {
		utils.RespondJSON(w, http.StatusInternalServerError, utils.Status{Message: "error occurred in parsing query"})
		return
	}
	if find {
		err := dbhelper.IncrementCartItem(cartId, productId, quantity)
		if err != nil {
			utils.RespondJSON(w, http.StatusInternalServerError, utils.Status{Message: "error occurred in updating of cart item"})
			return
		}
	} else {
		err := dbhelper.InsertCartItem(cartId, productId, 1)
		if err != nil {
			utils.RespondJSON(w, http.StatusInternalServerError, utils.Status{Message: "error occurred in query parsing"})
			return
		}
	}
	utils.RespondJSON(w, http.StatusOK, utils.Status{Message: "item added successfully in cart"})
}

func DeleteCartItem(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("claims").(*models.Claims)
	userId := claims.UserId
	cartId, CartErr := dbhelper.CartIdFromUserId(userId)
	if CartErr != nil {
		utils.RespondJSON(w, http.StatusInternalServerError, utils.Status{Message: "error occurred in parsing table"})
		return
	}
	productIdString := r.URL.Query().Get("id")
	productId, parseErr := uuid.Parse(productIdString)
	if parseErr != nil {
		utils.RespondJSON(w, http.StatusInternalServerError, utils.Status{Message: "error occurred into parsing of uuid"})
		return
	}
	_, quantity, err := dbhelper.FindProductFromCartItems(productId, cartId)
	if err != nil {
		utils.RespondJSON(w, http.StatusInternalServerError, utils.Status{Message: "error occurred in parsing query"})
		return
	}
	if quantity == 1 {
		err := dbhelper.DeleteCartItem(cartId, productId)
		if err != nil {
			utils.RespondJSON(w, http.StatusInternalServerError, utils.Status{Message: "error occurred in query parsing"})
			return
		}
	} else {
		err := dbhelper.DecrementCartItem(cartId, productId, quantity)
		if err != nil {
			utils.RespondJSON(w, http.StatusInternalServerError, utils.Status{Message: "error occurred in updating of cart item"})
			return
		}
	}
}

func ShowCartItems(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("claims").(*models.Claims)
	userId := claims.UserId
	cartId, CartErr := dbhelper.CartIdFromUserId(userId)
	if CartErr != nil {
		utils.RespondJSON(w, http.StatusInternalServerError, utils.Status{Message: "error occurred in parsing table"})
		return
	}
	cartItems, err := dbhelper.ShowCartItems(cartId)
	if err != nil {
		utils.RespondJSON(w, http.StatusInternalServerError, utils.Status{Message: "error occurred in parsing query"})
		return
	}
	utils.RespondJSON(w, http.StatusOK, cartItems)
}
