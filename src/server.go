package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"os"
	_ "github.com/mattn/go-sqlite3"
)

type err_values struct { 
    Code int 
    Msg string
}

type task_t struct { 
    Id int;
    Task string;
}

var db *sql.DB;
var t *template.Template;

func main()  {
    db_init();
    template_init();
    assets_init();

    http.HandleFunc("/change_priority", change_priority);
    http.HandleFunc("/delete_task", delete_task); 
    http.HandleFunc("/logout", logout);
    http.HandleFunc("/home", home);
    http.HandleFunc("/register", register);
    http.HandleFunc("/login", login);
    http.HandleFunc("/", root);

    err := http.ListenAndServe(":8080", nil);
    if err != nil {
        println("ERROR: opening port");
        log.Fatal(err)
    }
}

func unexpected_err(w http.ResponseWriter, err error, msg string, args...any) {
    log.Printf(msg, args)
    log.Println(err)
    w.WriteHeader(501)
    err = t.ExecuteTemplate(w, "error.html", err_values{501, 
                                    "Internal Server Error"})
    if err != nil{
        println("ERROR: error.html render")
        log.Println(err)
    }
}

func create_session(w http.ResponseWriter,  user_id int) (string, error) { 
    uuid, err := os.ReadFile("/proc/sys/kernel/random/uuid")
    if err != nil {
        unexpected_err(w, err, 
            "ERROR: generating uuid for session_id error")
        return "nil", err
    }
    session_id := string(uuid[:len(uuid)-1]);
    db.Exec("INSERT INTO sessions(user_id, token) values(?, ?)", 
            user_id, session_id);

    return session_id, nil;
}

func template_init() {
    var err error;
    t, err = template.ParseFiles(
        "views/index.html",
        "views/main.html",
        "views/error.html",
        "views/register.html",
        "views/login.html",
        "views/home.html",
    );
    if err != nil {
        println("ERROR: parsing html files")
        log.Fatal(err) 
    }
}

func db_init() {
    var err error;
    db, err = sql.Open("sqlite3", "./database.db")
    if err != nil {
        println("ERROR: opening database")
        log.Fatal(err) 
    }

    _, err = db.Exec(`
        PRAGMA  foreign_keys = ON; 
        CREATE TABLE IF NOT EXISTS users( 
            id INTEGER PRIMARY KEY AUTOINCREMENT, 
            username TEXT NOT NULL UNIQUE, 
            password TEXT NOT NULL 
        );

        CREATE TABLE IF NOT EXISTS sessions(
            id INTEGER PRIMARY KEY AUTOINCREMENT, 
            user_id INTEGER NOT NULL,
            token TEXT NOT NULL UNIQUE,
            FOREIGN KEY (user_id) REFERENCES users(id)
        );

        CREATE TABLE IF NOT EXISTS todo(
            id INTEGER PRIMARY KEY AUTOINCREMENT, 
            user_id INTEGER NOT NULL,
            task TEXT,
            FOREIGN KEY (user_id) REFERENCES users(id)
        );

        CREATE TABLE IF NOT EXISTS task_priority(
            id INTEGER PRIMARY KEY AUTOINCREMENT, 
            task_id INTEGER,
            priority INTEGER,
            FOREIGN KEY (task_id) REFERENCES todo(id) ON DELETE CASCADE
        );
    `);
    if err != nil {
        println("ERROR: creaing tables")
        db.Close()
        log.Fatal(err)
    }
}

func assets_init()  {
    ret, err := os.ReadDir("assets")
    if err != nil {
        log.Fatal(err)
    }

    for _, v := range(ret) { 
        if !v.IsDir() {
            name :=  v.Name();
            var type_header string;
            for i := len(name) - 1; i > 0; i-- {
                if name[i] == '.' {
                    switch name[i:] {
                        case ".js":
                            type_header = "application/javascript" 
                            break

                        case ".css":
                            type_header = "text/css" 
                            break
                    }
                    break
                }
            }

            name = "assets/" + name;
            // TODO: change once im sure of what i want
            // content, err := os.ReadFile(name);
            // if err != nil {
            //     log.Fatal(err)
            // }

            http.HandleFunc("/" + name, func(w http.ResponseWriter, r *http.Request) {
                w.Header().Set("Content-Type", type_header);
                content, _:= os.ReadFile(name);
                w.Write(content)
            })
        }
    }
}

