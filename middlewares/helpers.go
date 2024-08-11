package middlewares

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
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

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(user)

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

	user.ProfilePicture = &picRes.Results[0].Pictures.Large

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
			"message":         "user created",
			"success":         true,
			"token":           user.Uuid,
			"username":        user.Username,
			"name":            user.Name,
			"profile_picture": user.ProfilePicture,
		},
	)

	return nil

}

// login user
type comingUserLogin struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (r *Repository) LoginUser(context *fiber.Ctx) error {
	user := comingUserLogin{}
	dbResUser := models.Users{}

	// Log the incoming request body for debugging
	body := context.Body()
	log.Printf("Incoming request body: %s", body)

	err := context.BodyParser(&user)
	if err != nil {
		log.Printf("Error parsing body: %v", err)
		return context.Status(http.StatusUnprocessableEntity).JSON(&fiber.Map{
			"message": "request failed",
			"error":   err.Error(),
		})
	}

	// Log the parsed user struct for debugging
	log.Printf("Parsed user struct: %+v", user)
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

	fmt.Printf("%+v", dbResUser)

	context.Status(http.StatusOK).JSON(
		&fiber.Map{
			"success":         true,
			"token":           dbResUser.Uuid,
			"username":        dbResUser.Username,
			"name":            dbResUser.Name,
			"profile_picture": dbResUser.ProfilePicture,
		},
	)
	return nil
}

// get user data
func (r *Repository) GetUserData(context *fiber.Ctx) error {
	searchedUser := models.Users{}

	checkLoggedInUser, _ := r.IsAuthUser(context)

	if !checkLoggedInUser {
		context.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"message": "un-authorized"})
		return nil
	}

	err := r.DB.Where("uuid = ?", context.Params("user_id")).First(&searchedUser).Error

	if err != nil {
		context.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"message": "user not found"})
		return err
	}

	context.Status(http.StatusOK).JSON(&fiber.Map{
		"success": true,
		"data":    searchedUser,
	})
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

	err = r.DB.Preload("Classroom").Where("user_id = ? AND role = ? AND is_removed = ?", user.Uuid, "student", false).Find(&joinedClassroom).Error

	// if err != nil {
	// 	context.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
	// 		"message": "could not get joined classrooms",
	// 		"success": false,
	// 	})
	// 	return err
	// }

	err = r.DB.Preload("Classroom").Where("user_id = ? AND role = ? AND is_removed = ?", user.Uuid, "teacher", false).Find(&joinedTeacherClassroom).Error

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

// list a single classrooms
func (r *Repository) GetSingleClassroom(context *fiber.Ctx) error {

	var Classroom models.Classroom

	checkUserLoggedIn, _ := r.IsAuthUser(context)

	if !checkUserLoggedIn {
		context.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
			"message": "un-authorized",
			"success": false,
		})
		return nil
	}

	classid := context.Params("class_id")

	err := r.DB.Preload("Owner").Where("class_id = ? AND is_deleted = ?", classid, false).First(&Classroom).Error
	fmt.Println(err)

	context.Status(http.StatusOK).JSON(&fiber.Map{
		"success": true,
		"data":    Classroom,
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
	foundData := []models.ClassroomCollaborator{}
	err = r.DB.Where("class_id = $1 AND user_id = $2", collaborator.ClassID, user.Uuid).Find(&foundData).Error

	if len(foundData) != 0 {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"success": false, "message": "you are already a member of this classroom"})
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
	checkUserLoggedIn, _ := r.IsAuthUser(context)

	if !checkUserLoggedIn {
		context.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
			"message": "un-authorized",
			"success": false,
		})
		return nil
	}

	classId := context.Params("class_id")
	userId := context.Params("user_id")

	collaborator := models.ClassroomCollaborator{}

	err := r.DB.Where("user_id = ? AND class_id = ?", userId, classId).First(&collaborator).Error

	if err != nil {
		context.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
			"message": "member not found",
			"success": false,
		})
		return err
	}

	err = r.DB.Model(&collaborator).Where("user_id = ? AND class_id = ?", userId, classId).Update("is_removed", true).Error

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
	classDetails := models.Classroom{}

	checkUserLoggedIn, _ := r.IsAuthUser(context)

	if !checkUserLoggedIn {
		context.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
			"message": "un-authorized",
			"success": false,
		})
		return nil
	}

	err := r.DB.Where("class_id = ? ", context.Params("class_id")).First(&classDetails).Error

	if err != nil {
		context.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
			"message": "could not get classroom",
			"success": false,
		})
		return err
	}
	err = r.DB.Preload("User").Where("class_id = ? AND role = ? AND is_removed = ?", context.Params("class_id"), "student", false).Find(&clasroomStudents).Error
	fmt.Println(err)
	err = r.DB.Preload("User").Where("class_id = ? AND role = ? AND is_removed = ?", context.Params("class_id"), "teacher", false).Find(&clasroomTeachers).Error

	context.Status(fiber.StatusOK).JSON(&fiber.Map{
		"owner_id": classDetails.OwnerID,
		"message":  "students fetched",
		"success":  true,
		"students": clasroomStudents,
		"teachers": clasroomTeachers,
	})

	return nil
}

/*-----------------------------------------------------assignment section----------------------------------------------------------------*/

func (r *Repository) CreateAssignment(context *fiber.Ctx) error {
	checkUserLoggedIn, user := r.IsAuthUser(context)

	if !checkUserLoggedIn {
		context.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
			"message": "un-authorized",
			"success": false,
		})
		return nil
	}

	incomingAssignment := models.Assignments{}

	err := context.BodyParser(&incomingAssignment)

	if err != nil {
		context.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
			"message": "request failed",
			"success": false,
		})
		return err
	}

	id, _ := utils.GenerateUUid()

	incomingAssignment.ID = &id
	incomingAssignment.AutherId = user.Uuid

	err = r.DB.Create(&incomingAssignment).Error

	if err != nil {
		context.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
			"message": "database insertion failed",
			"success": false,
		})
		return err
	}

	context.Status(fiber.StatusOK).JSON(&fiber.Map{
		"message": "assignment created",
		"success": true,
		"data":    incomingAssignment,
	})

	return nil
}
func (r *Repository) EditAssignment(context *fiber.Ctx) error {
	checkUserLoggedIn, user := r.IsAuthUser(context)

	if !checkUserLoggedIn {
		context.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
			"message": "un-authorized",
			"success": false,
		})
		return nil
	}

	incomingAssignment := models.Assignments{}

	err := context.BodyParser(&incomingAssignment)

	incomingAssignmentId := context.Params("id")

	if err != nil {
		context.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
			"message": "request failed",
			"success": false,
		})
		return err
	}

	dbResAssignment := models.Assignments{}

	err = r.DB.Where("id = ?", incomingAssignmentId).First(&dbResAssignment).Error

	if dbResAssignment.ID == nil {
		context.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
			"message": "assignment not found",
			"success": false,
		})
		return nil
	}

	if *dbResAssignment.AutherId != *user.Uuid {
		context.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
			"message": "un-authorized",
			"success": false,
		})
		return nil
	}
	err = r.DB.Model(&dbResAssignment).Updates(&incomingAssignment).Error

	if err != nil {
		context.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
			"message": "database update failed",
			"success": false,
		})
		return err
	}

	context.Status(fiber.StatusOK).JSON(&fiber.Map{
		"message": "assignment edited",
		"success": true,
	})

	return nil
}

func (r *Repository) GetAllAssignments(context *fiber.Ctx) error {
	isUserLoggedIn, _ := r.IsAuthUser(context)

	if !isUserLoggedIn {
		context.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
			"message": "un-authorized",
			"success": false,
		})
		return nil
	}

	incomingClassId := context.Params("class_id")

	allAssignments := []models.Assignments{}

	err := r.DB.Preload("Classroom").Preload("CreatedBy").Where("class_id = ? AND is_deleted = ?", incomingClassId, false).Find(&allAssignments).Error

	if err != nil {
		context.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
			"message": "could not get assignments",
			"success": false,
		})
		return err
	}
	context.Status(fiber.StatusOK).JSON(&fiber.Map{
		"message": "assignments fetched",
		"success": true,
		"data":    allAssignments,
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

	// Default middleware config allows all origins
	app.Use(cors.New())

	// Middleware configuration to allow all origins without credentials
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*", // Allows all origins
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
		AllowHeaders:     "Content-Type, Authorization",
		AllowCredentials: false, // Do not allow credentials
	}))

	/*---------------------user routes----------------------*/
	api.Post("/user/create", r.CreateUser)
	api.Post("/user/login", r.LoginUser)
	api.Post("/user/:user_id", r.GetUserData)

	/*---------------------classroom routes----------------------*/
	api.Post("/classroom/create", r.CreateClassroom)
	api.Get("/classrooms", r.GetClassrooms)
	api.Get("/classroom/:class_id", r.GetSingleClassroom)
	api.Patch("/classroom/edit/:class_id", r.EditClassroom)
	api.Post("/classroom/join", r.JoinClassroom)
	api.Patch("/classroom/exit/:class_id/:user_id", r.ExitClassroom)
	api.Get("/classroom/members/:class_id", r.ListAllMembers)

	/*-----------------------assignment routes----------------------*/

	api.Post("/assignment/create", r.CreateAssignment)
	api.Patch("/assignment/:id/edit", r.EditAssignment)
	api.Get("/assignments/:class_id", r.GetAllAssignments)

	api.Get("/test", r.testMessage)
}
