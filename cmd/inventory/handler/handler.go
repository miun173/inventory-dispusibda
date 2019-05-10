package handler

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/gorilla/mux"
	"github.com/miun173/inventory-dispusibda/cmd/inventory/models"
	"github.com/miun173/inventory-dispusibda/cmd/inventory/repo"
)

func decodeAuthToken(authToken string) (int, string, error) {
	s := strings.SplitN(authToken, " ", 2)
	if len(s) != 2 {
		return 0, "", errors.New("authToken not valid")
	}

	b, err := base64.StdEncoding.DecodeString(s[1])
	if err != nil {
		return 0, "", errors.WithStack(err)
	}

	pair := strings.SplitN(string(b), ":", 2)
	if len(pair) != 2 {
		return 0, "", errors.New("authToken not valid")
	}

	id, err := strconv.Atoi(pair[0])
	if err != nil {
		return 0, "", errors.WithStack(err)
	}

	return id, pair[1], nil
}

// CheckAuth validate user auth
func CheckAuth(w http.ResponseWriter, r *http.Request) {
	var user models.User
	_ = json.NewDecoder(r.Body).Decode(&user)

	authToken := r.Header.Get("Authorization")
	id, token, err := decodeAuthToken(authToken)
	if err != nil {
		log.Printf("%+v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var userRepo models.User
	err = repo.GetUserByID(&userRepo, id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if token != userRepo.Token || user.Role != userRepo.Role {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// Ping check if service ready
func Ping(w http.ResponseWriter, r *http.Request) {
	msg := map[string]string{"message": "pong"}
	log.Print("ping")
	json.NewEncoder(w).Encode(msg)
}

// CreateUser handle user creation
func CreateUser(w http.ResponseWriter, r *http.Request) {
	var user models.User
	_ = json.NewDecoder(r.Body).Decode(&user)
	err := repo.CreateUser(&user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "system error"})
		return
	}

	json.NewEncoder(w).Encode(user)
}

// GetAllUsers handle get all users
func GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := repo.GetAllUser()
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": "system error"})
		return
	}

	json.NewEncoder(w).Encode(users)
}

// Login handle user login
func Login(w http.ResponseWriter, r *http.Request) {
	var user models.User
	_ = json.NewDecoder(r.Body).Decode(&user)

	var userRepo models.User
	err := repo.GetUserByUsername(&userRepo, user.Username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "system error"})
		return
	}

	if userRepo.ID == 0 || userRepo.Password != user.Password {
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(map[string]string{"error": "incorect username or password"})
		return
	}

	userRepo.Password = ""
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(userRepo)
}

// CreateBarang handle create barang
func CreateBarang(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// barang diambil dari inputan
	var brg models.Barang
	_ = json.NewDecoder(r.Body).Decode(&brg)

	// barang dimasukkan database
	err := repo.CreateBarang(&brg)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": "system error"})
		return
	}

	json.NewEncoder(w).Encode(brg)
}

// GetBarang handle get all barang
func GetBarang(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Fatal(err)
		return
	}

	var brg models.Barang
	err = repo.GetBarang(id, &brg)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": "system error"})
		return
	}

	json.NewEncoder(w).Encode(brg)
}

// CreateBarangKeluar handle barang keluar
func CreateBarangKeluar(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var brg models.BarangKeluar
	_ = json.NewDecoder(r.Body).Decode(&brg)

	isBarangExists, err := repo.CheckBarangExists(brg.BarangID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "system error"})
		return
	}

	if !isBarangExists {
		errMsg := fmt.Sprintf("barang with barangID %d is not exists", brg.BarangID)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": errMsg})
		return
	}

	// create barang if barangID exist
	err = repo.CreateBarangKeluar(&brg)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "system error"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(brg)
}

// GetAllBarangKeluar list all available barang keluar
func GetAllBarangKeluar(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	brgs, err := repo.GetAllBarangKeluar()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "system error"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(brgs)
}

// GetAllBarang list all available barang keluar
func GetAllBarang(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	brgs, err := repo.GetAllBarang()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "system error"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(brgs)
}

// GetJurnal list all available jurnal/inventaris
func GetJurnal(w http.ResponseWriter, r *http.Request) {
	// inisiasi
	jurnals := make([]models.Jurnal, 0)
	brgs, err := repo.GetAllBarang()
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": "system error"})
		return
	}

	for _, b := range brgs {
		penyusutan := (b.Harga - b.NilaiSisa) / float64(b.UmurEkonomis)
		j := models.Jurnal{
			Barang:     b,
			Penyusutan: penyusutan,
		}
		jurnals = append(jurnals, j)
	}

	jData, err := json.Marshal(jurnals)
	if err != nil {
		log.Printf("error: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "something bad"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jData)
}
