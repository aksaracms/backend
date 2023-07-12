package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

// Database credentials
const (
	DBUsername = "root"
	DBPassword = ""
	DBName     = "weblat"
)

var db *sql.DB
var tpl *template.Template

type User struct {
	ID       int
	Name     string
	Email    string
	Username string
	Password string
	Role     string // 'admin' or 'user'
}

type Post struct {
	ID      int
	Title   string
	Content string
}

type Image struct {
	URL      string
	Filename string
}

type ContactEntry struct {
	ID      int
	Name    string
	Email   string
	Message string
}

func init() {
	// Connect to the database
	connStr := fmt.Sprintf("%s:%s@/%s", DBUsername, DBPassword, DBName)
	var err error
	db, err = sql.Open("mysql", connStr)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize the template
	tpl = template.Must(template.ParseGlob("templates/*.html"))
}

func main() {
	defer db.Close()

	// Routes
	http.HandleFunc("/home-usr", indexHandler)
	http.HandleFunc("/home-adm", homeHandler)
	http.HandleFunc("/posts", getPostsHandler)
	http.HandleFunc("/profile", getProfileHandler)
	http.HandleFunc("/gallery", galleryHandler)
	http.HandleFunc("/contact", contactHandler)
	http.HandleFunc("/contact/list", getContactListHandler)
	http.HandleFunc("/posts-admin", postsHandler)
	http.HandleFunc("/post/create", createPostHandler)
	http.HandleFunc("/post/edit", editPostHandler)
	http.HandleFunc("/post/delete", deletePostHandler)
	http.HandleFunc("/galery-admin", getImageHandler)
	http.HandleFunc("/galery/create", uploadImageHandler)
	http.HandleFunc("/galery/delete", deleteImageHandler)

	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/", loginHandler)
	http.HandleFunc("/logout", logoutHandler)

	log.Println("Server started on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tpl, err := template.ParseFiles("templates/landing.html")
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = tpl.Execute(w, nil)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func getPostsHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve posts from the database
	rows, err := db.Query("SELECT * FROM posts")
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.ID, &post.Title, &post.Content); err != nil {
			log.Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		posts = append(posts, post)
	}
	if err := rows.Err(); err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Render the posts
	tpl, err := template.ParseFiles("templates/posts.html")
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = tpl.Execute(w, posts)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func getProfileHandler(w http.ResponseWriter, r *http.Request) {
	// Define the user data
	user := User{
		Name:     "John Doe",
		Email:    "john@example.com",
		Username: "johndoe",
	}

	// Render the profile page template with the user data
	tpl, err := template.ParseFiles("templates/profile.html")
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = tpl.Execute(w, user)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func galleryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Handle form submissions or other POST requests here, if needed
		// ...

		// Redirect or display a success message
		http.Redirect(w, r, "/gallery", http.StatusSeeOther)
		return
	}

	// Retrieve image URLs from the "gallery" table in the database
	imageURLs, err := fetchImageURLsFromDatabase()
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Pass the image URLs to the template for rendering
	data := struct {
		ImageURLs []string
	}{
		ImageURLs: imageURLs,
	}

	// Render the gallery template with the image URLs
	tpl, err := template.ParseFiles("templates/gallery.html")
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = tpl.Execute(w, data)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func contactHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Parse the form data
		err := r.ParseForm()
		if err != nil {
			log.Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Retrieve the form values
		name := r.Form.Get("name")
		email := r.Form.Get("email")
		message := r.Form.Get("message")

		// Insert the form data into the database
		_, err = db.Exec("INSERT INTO contact_entries (name, email, message) VALUES (?, ?, ?)", name, email, message)
		if err != nil {
			log.Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Redirect or display a success message
		http.Redirect(w, r, "/success", http.StatusSeeOther)
		return
	}

	// Render the contact form template
	tpl, err := template.ParseFiles("templates/contact.html")
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = tpl.Execute(w, nil)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func getContactListHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve contact entries from the database
	rows, err := db.Query("SELECT * FROM contact_entries")
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var contactEntries []ContactEntry
	for rows.Next() {
		var entry ContactEntry
		if err := rows.Scan(&entry.ID, &entry.Name, &entry.Email, &entry.Message); err != nil {
			log.Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		contactEntries = append(contactEntries, entry)
	}
	if err := rows.Err(); err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Render the contact list template
	tpl, err := template.ParseFiles("templates/contact_list.html")
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = tpl.Execute(w, contactEntries)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func fetchImageURLsFromDatabase() ([]string, error) {
	// Execute a query to fetch image URLs from the "gallery" table
	rows, err := db.Query("SELECT imageURL FROM gallery")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Iterate over the rows and extract the imageURL column
	var imageURLs []string
	for rows.Next() {
		var imageURL string
		err := rows.Scan(&imageURL)
		if err != nil {
			return nil, err
		}
		imageURLs = append(imageURLs, imageURL)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return imageURLs, nil
}

func postsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Fetch all posts from the database
		posts, err := fetchPostsFromDatabase()
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Render the posts page with the list of posts
		tpl, err := template.ParseFiles("templates/postsadm.html")
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		err = tpl.Execute(w, posts)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	} else {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func createPostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Render the create post form
		tpl, err := template.ParseFiles("templates/create_post.html")
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		err = tpl.Execute(w, nil)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	} else if r.Method == http.MethodPost {
		// Retrieve the form data
		title := r.FormValue("title")
		content := r.FormValue("content")

		// Save the new post to the database
		err := savePostToDatabase(title, content)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Redirect to the posts page or display a success message
		http.Redirect(w, r, "/posts", http.StatusSeeOther)
	} else {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func editPostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Retrieve the post ID from the query parameters
		postID := r.URL.Query().Get("id")

		// Fetch the post from the database by ID
		post, err := fetchPostByID(postID)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Render the edit post form with the post data
		tpl, err := template.ParseFiles("templates/edit_post.html")
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		err = tpl.Execute(w, post)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	} else if r.Method == http.MethodPost {
		// Retrieve the form data
		postID := r.FormValue("id")
		title := r.FormValue("title")
		content := r.FormValue("content")

		// Update the post in the database
		err := updatePostInDatabase(postID, title, content)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Redirect to the posts page or display a success message
		http.Redirect(w, r, "/posts", http.StatusSeeOther)
	} else {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func deletePostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Retrieve the post ID from the form data
		postID := r.FormValue("id")

		// Delete the post from the database
		err := deletePostFromDatabase(postID)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Redirect to the posts page or display a success message
		http.Redirect(w, r, "/posts", http.StatusSeeOther)
	} else {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func savePostToDatabase(title, content string) error {
	// Prepare the SQL statement
	stmt, err := db.Prepare("INSERT INTO posts (title, content) VALUES (?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Execute the SQL statement
	_, err = stmt.Exec(title, content)
	if err != nil {
		return err
	}

	return nil
}

func fetchPostsFromDatabase() ([]*Post, error) {
	// TODO: Implement the logic to fetch posts from the database
	// Prepare the SQL statement
	rows, err := db.Query("SELECT * FROM posts")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Create a slice to hold the retrieved posts
	posts := make([]*Post, 0)

	// Iterate over the rows
	for rows.Next() {
		// Create a new Post struct
		post := &Post{}

		// Scan the row values into the Post struct
		err := rows.Scan(&post.ID, &post.Title, &post.Content)
		if err != nil {
			return nil, err
		}

		// Append the post to the slice
		posts = append(posts, post)
	}

	// Check for any errors during iteration
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return posts, nil
}

func fetchPostByID(postID string) (*Post, error) {
	// Prepare the SQL statement
	stmt, err := db.Prepare("SELECT id, title, content FROM posts WHERE id = ?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	// Execute the SQL statement and retrieve the post
	var post Post
	err = stmt.QueryRow(postID).Scan(&post.ID, &post.Title, &post.Content)
	if err != nil {
		return nil, err
	}

	return &post, nil
}

func updatePostInDatabase(postID, title, content string) error {
	// Prepare the SQL statement
	stmt, err := db.Prepare("UPDATE posts SET title = ?, content = ? WHERE id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Execute the SQL statement
	_, err = stmt.Exec(title, content, postID)
	if err != nil {
		return err
	}

	return nil
}

func deletePostFromDatabase(postID string) error {
	// Prepare the SQL statement
	stmt, err := db.Prepare("DELETE FROM posts WHERE id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Execute the SQL statement
	_, err = stmt.Exec(postID)
	if err != nil {
		return err
	}

	return nil
}

func getImageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Handle form submissions or other POST requests here, if needed
		// ...

		// Redirect or display a success message
		http.Redirect(w, r, "/galery-admin", http.StatusSeeOther)
		return
	}

	// Retrieve image URLs from the "gallery" table in the database
	imageURLs, err := fetchImageURLsFromDatabase()
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Pass the image URLs to the template for rendering
	data := struct {
		ImageURLs []string
	}{
		ImageURLs: imageURLs,
	}
	fmt.Println(imageURLs)

	// Render the gallery template with the image URLs
	tpl, err := template.ParseFiles("templates/galeryadm.html")
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = tpl.Execute(w, data)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func deleteImageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Retrieve the image URL from the form data
		imageURL := r.FormValue("imageURL")

		// Delete the image from the "gallery" table in the database
		_, err := db.Exec("DELETE FROM gallery WHERE imageURL = ?", imageURL)
		if err != nil {
			log.Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Delete the image file from the "uploads" directory
		// filename := imageURL[len("/uploads/"):]
		// err = os.Remove("./uploads/" + filename)
		// if err != nil {
		// 	log.Println(err)
		// 	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		// 	return
		// }

		// Redirect or display a success message
		http.Redirect(w, r, "/gallery", http.StatusSeeOther)
		return
	}

	http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
}

func uploadImageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Retrieve the uploaded file
		file, handler, err := r.FormFile("file")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Save the uploaded file with a unique filename
		filename := handler.Filename
		f, err := os.OpenFile("uploads/"+filename, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer f.Close()
		io.Copy(f, file)

		// Update the profile picture URL in the database
		imageURL := "uploads/" + filename
		_, err = db.Exec("INSERT INTO gallery (imageURL) VALUES (?)", imageURL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Render the index page with the updated image gallery
		imageURLs, err := fetchImageURLsFromDatabase()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tpl, err := template.ParseFiles("templates/index.html")
		if err != nil {
			log.Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		err = tpl.Execute(w, imageURLs)
		if err != nil {
			log.Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	// Render the upload image form
	tpl, err := template.ParseFiles("templates/upload-image.html")
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = tpl.Execute(w, nil)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	tpl, err := template.ParseFiles("templates/dashboard.html")
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = tpl.Execute(w, nil)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Get the form values
		name := r.FormValue("name")
		email := r.FormValue("email")
		username := r.FormValue("username")
		password := r.FormValue("password")

		// Create a new user
		user := User{
			Name:     name,
			Email:    email,
			Username: username,
			Password: password,
			Role:     "user",
		}

		// Save the user to the database
		stmt, err := db.Prepare("INSERT INTO users (name, username, email, password, role) VALUES (?, ?, ?, ?, ?)")
		if err != nil {
			log.Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(user.Name, user.Username, user.Email, user.Password, user.Role)
		if err != nil {
			log.Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Redirect to the login page
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	err := tpl.ExecuteTemplate(w, "register.html", nil)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")

		// Fetch the user from the database
		user, err := fetchUserByUsername(username, db)
		if err != nil {
			http.Error(w, "Error fetching user", http.StatusInternalServerError)
			return
		}

		// Check if the user exists and password is correct
		if user == nil || user.Password != password {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		// If login is successful, redirect to appropriate pages based on user role
		if user.Role == "admin" {
			http.Redirect(w, r, "/home-adm", http.StatusFound)
		} else if user.Role == "user" {
			http.Redirect(w, r, "/home-usr", http.StatusFound)
		} else {
			http.Error(w, "Invalid user role", http.StatusInternalServerError)
		}
	} else {
		// Display the login form
		tpl.ExecuteTemplate(w, "login.html", nil)
	}
}

func fetchUserByUsername(username string, db *sql.DB) (*User, error) {
	query := "SELECT id, name, email, username, password, role FROM users WHERE username = ?"
	row := db.QueryRow(query, username)

	var user User
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Username, &user.Password, &user.Role)
	if err != nil {
		if err == sql.ErrNoRows {
			// User not found
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	// Clear the session cookie to log out the user
	cookie := &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(w, cookie)
	// Redirect to the login page or any other desired page
	http.Redirect(w, r, "/", http.StatusFound)
}

func validateCredentials(username, password string) bool {
	// Query the database for the user with the given username and password
	row := db.QueryRow("SELECT COUNT(*) FROM users WHERE username = ? AND password = ?", username, password)
	var count int
	err := row.Scan(&count)
	if err != nil {
		log.Println(err)
		return false
	}

	// Return true if a user record with the provided username and password exists, otherwise return false
	return count > 0
}

func authenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the user is authenticated
		authenticated := checkAuthentication(r)
		if !authenticated {
			// If not authenticated, redirect to the login page or display an error message
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		// If authenticated, call the next handler
		next.ServeHTTP(w, r)
	})
}

func checkAuthentication(r *http.Request) bool {
	// Check if the session cookie is present and contains the authenticated value
	cookie, err := r.Cookie("session")
	if err != nil || cookie.Value != "authenticated" {
		return false
	}
	return true
}
