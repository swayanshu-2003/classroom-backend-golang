package middlewares

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/swayanshu-2003/classroom-backend/models"
	"github.com/swayanshu-2003/classroom-backend/utils"

	"gorm.io/gorm"
)

type Repository struct {
	DB *gorm.DB
}

type picRes struct {
	Results []Result `json:"results"`
}

type Result struct {
	Pictures Picture `json:"picture"`
}

type Picture struct {
	Large     string `json:"large"`
	Medium    string `json:"medium"`
	Thumbnail string `json:"thumbnail"`
}

// check if user is logged in or not
func (r *Repository) IsAuthUser(context *fiber.Ctx) (bool, *models.Users) {
	userId := context.Get("authorization")

	if len(userId) == 0 {
		return false, nil
	}

	var user models.Users
	err := r.DB.Where("uuid = ?", userId).First(&user).Error
	if err != nil {
		return false, nil
	}

	return true, &user
}

/*------------------------------------------------ user helpers ------------------------------------------------------*/
//register user
func (r *Repository) CreateUser(context *fiber.Ctx) error {
	user := models.Users{}

	err := context.BodyParser(&user)

	uuid, _ := utils.GenerateUUid()

	user.Uuid = &uuid

	res, _ := http.Get("https://randomuser.me/api/?inc=picture&results=1")

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}

	var picRes picRes

	// _ = json.NewDecoder(res.Body).Decode(&picRes)

	err = json.Unmarshal(body, &picRes)
	if err != nil {
		log.Fatalf("Error unmarshalling JSON: %v", err)
	}

	user.ProfilePicture = picRes.Results[0].Pictures.Large

	fmt.Println(user)

	if err != nil {
		context.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"message": "request failed"})
		return err
	}

	dbErr := r.DB.Create(&user).Error

	if dbErr != nil {
		context.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"message": "database insertion failed"})
		return dbErr
	}

	context.Status(http.StatusOK).JSON(
		&fiber.Map{
			"message":  "user created",
			"success":  true,
			"token":    user.Uuid,
			"username": user.Username,
			"name":     user.Name,
		},
	)

	return nil

}

// login user
func (r *Repository) LoginUser(context *fiber.Ctx) error {
	user := models.Users{}
	dbResUser := models.Users{}

	err := context.BodyParser(&user)

	if err != nil {
		context.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"message": "request failed"})
		return err
	}
	// if err != nil {
	// 	context.Status(http.StatusUnprocessableEntity).JSON(
	// 		&fiber.Map{"message": "request failed"})
	// 	return err
	// }

	err = r.DB.Where("username = ?", user.Username).First(&dbResUser).Error

	if err != nil {
		context.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"message": "invalid username"})
		return err
	}

	err = r.DB.Where("username = ? AND password >= ?", user.Username, user.Password).Find(&dbResUser).Error

	if err != nil {
		context.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"message": "invalid invalid password"})
		return err
	}

	context.Status(http.StatusUnprocessableEntity).JSON(
		&fiber.Map{
			"success":  true,
			"token":    dbResUser.Uuid,
			"username": dbResUser.Username,
			"name":     dbResUser.Name,
		},
	)
	return nil
}

/*------------------------------------------------ classroom helpers ------------------------------------------------------*/
//create classroom
func (r *Repository) CreateClassroom(context *fiber.Ctx) error {
	classroom := models.Classroom{}

	checkLoggedInUser, user := r.IsAuthUser(context)

	if !checkLoggedInUser {
		context.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"message": "un-authorized"})
		return nil
	}

	err := context.BodyParser(&classroom)

	if err != nil {
		context.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"message": "request failed"})
		return err
	}

	classroom.OwnerID = user.Uuid

	classId := utils.GenerateClassroomId()

	classroom.ClassId = classId

	fmt.Println(classroom)

	dbErr := r.DB.Create(&classroom).Error

	if dbErr != nil {
		context.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"message": "database insertion failed"})
		return dbErr
	}

	var collaborator models.ClassroomCollaborator

	collaborator.ClassID = classId
	collaborator.UserID = user.Uuid
	collaborator.Role = "teacher"

	dbErr = r.DB.Create(&collaborator).Error

	if dbErr != nil {
		context.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"message": "database insertion failed"})
		return dbErr
	}

	context.Status(http.StatusOK).JSON(&fiber.Map{
		"message":   "classroom created",
		"success":   true,
		"classroom": classroom,
	})
	return nil
}

// list all classrooms
func (r *Repository) GetClassrooms(context *fiber.Ctx) error {

	var Classrooms []models.Classroom

	joinedClassroom := []models.ClassroomCollaborator{}
	joinedTeacherClassroom := []models.ClassroomCollaborator{}

	checkUserLoggedIn, user := r.IsAuthUser(context)

	if !checkUserLoggedIn {
		context.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
			"message": "un-authorized",
			"success": false,
		})
		return nil
	}

	err := r.DB.Preload("Collaborators").Where("owner_id = ? AND is_deleted = ?", user.Uuid, false).Find(&Classrooms).Error
	fmt.Println(err)
	// if err != nil {
	// 	context.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
	// 		"message": "could not get owned classrooms",
	// 		"success": false,
	// 	})
	// 	return err
	// }

	err = r.DB.Where("user_id = ? AND role = ? AND is_removed = ?", user.Uuid, "student", false).Find(&joinedClassroom).Error

	// if err != nil {
	// 	context.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
	// 		"message": "could not get joined classrooms",
	// 		"success": false,
	// 	})
	// 	return err
	// }

	err = r.DB.Where("user_id = ? AND role = ? AND is_removed = ?", user.Uuid, "teacher", false).Find(&joinedTeacherClassroom).Error

	// if err != nil {
	// 	context.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
	// 		"message": "could not get joined classrooms",
	// 		"success": false,
	// 	})
	// 	return err
	// }

	context.Status(http.StatusOK).JSON(&fiber.Map{
		"success":           true,
		"own":               Classrooms,
		"joined_as_student": joinedClassroom,
		"joined_as_teacher": joinedTeacherClassroom,
	})
	return nil
}

// edit a class
func (r *Repository) EditClassroom(context *fiber.Ctx) error {
	classroom := models.Classroom{}

	checkLoggedInUser, user := r.IsAuthUser(context)

	if !checkLoggedInUser {
		context.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{
				"success": false,
				"message": "Please Login"})
		return nil
	}

	err := context.BodyParser(&classroom)

	if err != nil {
		context.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"success": false, "message": "request failed"})
		return err
	}

	class := models.Classroom{}

	err = r.DB.Where("class_id = ?", context.Params("class_id")).First(&class).Error

	if err != nil {
		context.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"success": false, "message": "classroom not found"})
		return err
	}

	if *class.OwnerID != *user.Uuid {
		context.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"success": false, "message": "un-authorized"})
		return nil
	}

	err = r.DB.Model(&class).Updates(classroom).Error

	if err != nil {
		context.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"success": false, "message": "database update failed"})
		return err
	}

	context.Status(http.StatusOK).JSON(&fiber.Map{
		"success": true,
		"message": "classroom updated",
		"data":    class,
	})

	return nil
}

// join classroom
func (r *Repository) JoinClassroom(context *fiber.Ctx) error {
	var collaborator models.ClassroomCollaborator

	checkLoggedInUser, user := r.IsAuthUser(context)

	if !checkLoggedInUser {
		context.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{
				"success": false,
				"message": "Please Login"})
		return nil
	}

	err := context.BodyParser(&collaborator)

	if err != nil {
		context.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"success": false, "message": "request failed"})
		return err
	}

	collaborator.UserID = user.Uuid

	dbErr := r.DB.Create(&collaborator).Error

	if dbErr != nil {
		context.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"message": "database insertion failed"})
		return dbErr
	}

	context.Status(http.StatusOK).JSON(&fiber.Map{
		"message": "Successfully Enrolled",
		"success": true,
		"data":    collaborator,
	})
	return nil
}

// exit or remove from classroom
func (r *Repository) ExitClassroom(context *fiber.Ctx) error {
	checkUserLoggedIn, user := r.IsAuthUser(context)

	if !checkUserLoggedIn {
		context.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
			"message": "un-authorized",
			"success": false,
		})
		return nil
	}

	classId := context.Params("class_id")

	collaborator := models.ClassroomCollaborator{}

	err := r.DB.Where("user_id = ? AND class_id = ?", user.Uuid, classId).First(&collaborator).Error

	if err != nil {
		context.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
			"message": "collaborator not found",
			"success": false,
		})
		return err
	}

	err = r.DB.Model(&collaborator).Where("user_id = ? AND class_id = ?", user.Uuid, classId).Update("is_removed", true).Error

	if err != nil {
		context.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
			"message": "database update failed",
			"success": false,
		})
		return err
	}

	context.Status(fiber.StatusOK).JSON(&fiber.Map{
		"message": "Successfully Exited",
		"success": true,
		"data":    collaborator,
	})

	return nil
}

// list allstudents
func (r *Repository) ListAllMembers(context *fiber.Ctx) error {
	clasroomStudents := []models.ClassroomCollaborator{}
	clasroomTeachers := []models.ClassroomCollaborator{}

	checkUserLoggedIn, _ := r.IsAuthUser(context)

	if !checkUserLoggedIn {
		context.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
			"message": "un-authorized",
			"success": false,
		})
		return nil
	}

	err := r.DB.Preload("User").Where("class_id = ? AND role = ?", context.Params("class_id"), "student").Find(&clasroomStudents).Error
	fmt.Println(err)
	// if err != nil {
	// 	context.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
	// 		"message": "could not get students",
	// 		"success": false,
	// 	})
	// 	return err
	// }
	err = r.DB.Preload("User").Where("class_id = ? AND role = ?", context.Params("class_id"), "teacher").Find(&clasroomTeachers).Error

	context.Status(fiber.StatusOK).JSON(&fiber.Map{
		"message":  "students fetched",
		"success":  true,
		"students": clasroomStudents,
		"teachers": clasroomTeachers,
	})

	return nil
}

/*------------------------------------------------ test ------------------------------------------------------*/

func (r *Repository) testMessage(context *fiber.Ctx) error {
	userId := context.Get("authorization")
	// userId := context.Get("test")
	fmt.Println(userId)

	res, _ := http.Get("https://randomuser.me/api/?inc=picture&results=1")

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}

	fmt.Println(string(body))

	var picRes picRes

	// _ = json.NewDecoder(res.Body).Decode(&picRes)

	err = json.Unmarshal(body, &picRes)
	if err != nil {
		log.Fatalf("Error unmarshalling JSON: %v", err)
	}

	fmt.Println(picRes)

	context.Status(http.StatusOK).JSON(&fiber.Map{"message": "test message", "user-id": userId, "data": picRes})

	return nil
}

func (r *Repository) SetupRoutes(app *fiber.App) {
	api := app.Group("/api/v1")

	/*---------------------user routes----------------------*/
	api.Post("/user/create", r.CreateUser)
	api.Post("/user/login", r.LoginUser)

	/*---------------------classroom routes----------------------*/
	api.Post("/classroom/create", r.CreateClassroom)
	api.Get("/classrooms", r.GetClassrooms)
	api.Patch("/classroom/edit/:class_id", r.EditClassroom)
	api.Post("/classroom/join", r.JoinClassroom)
	api.Patch("/classroom/exit/:class_id", r.ExitClassroom)
	api.Get("/classroom/members/:class_id", r.ListAllMembers)

	api.Get("/test", r.testMessage)
}
