package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/skip2/go-qrcode"
	"github.com/xuri/excelize/v2"
)

// User ساختار داده‌ای پرسنل
type User struct {
	FullName    string
	ChildCount  string
	SpouseName  string
	ChildName   string
	Landline    string
	MobilePhone string
}

var usersList []User

// تابع خواندن اکسل (همان کد قبلی)
func loadUsersFromExcel(filePath string) error {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	sheetName := f.GetSheetName(0)
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return err
	}

	usersList = []User{}
	for i, row := range rows {
		if i == 0 {
			continue
		}
		user := User{}
		if len(row) > 0 {
			user.FullName = row[0]
		}
		if len(row) > 1 {
			user.ChildCount = row[1]
		}
		if len(row) > 2 {
			user.SpouseName = row[2]
		}
		if len(row) > 3 {
			user.ChildName = row[3]
		}
		if len(row) > 4 {
			user.Landline = row[4]
		}
		if len(row) > 5 {
			user.MobilePhone = row[5]
		}

		if user.MobilePhone != "" {
			usersList = append(usersList, user)
		}
	}
	return nil
}

// تابع جستجوی کاربر بر اساس شماره تلفن همراه
func findUser(mobile string) (*User, bool) {
	// هر بار قبل از جستجو، اکسل را مجدد لود میکنیم تا اگر کاربری حذف شده بود اعمال شود
	_ = loadUsersFromExcel("users.xlsx")

	for _, u := range usersList {
		if u.MobilePhone == mobile {
			return &u, true
		}
	}
	return nil, false
}

// قالب HTML صفحه لاگین
const loginTemplate = `
<!DOCTYPE html>
<html lang="fa" dir="rtl">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>ورود به سیستم پرسنل</title>
    <!-- اضافه کردن فونت زیبای وزیرمتن برای ظاهر امروزی -->
    <link href="https://cdn.jsdelivr.net/gh/rastikerdar/vazirmatn@v33.003/Vazirmatn-font-face.css" rel="stylesheet" type="text/css" />
    <style>
        * {
            box-sizing: border-box;
            transition: all 0.3s ease;
        }
        body { 
            font-family: 'Vazirmatn', Tahoma, sans-serif; 
            background: linear-gradient(135deg, #e0eafc, #cfdef3); 
            display: flex; 
            justify-content: center; 
            align-items: center; 
            height: 100vh; 
            margin: 0; 
        }
        .login-box { 
            background: rgba(255, 255, 255, 0.95); 
            padding: 40px 30px; 
            border-radius: 16px; 
            box-shadow: 0 10px 25px rgba(0, 0, 0, 0.05); 
            width: 360px; 
            text-align: center; 
        }
        h3 {
            color: #2c3e50;
            margin-top: 0;
            margin-bottom: 25px;
            font-size: 20px;
            font-weight: 700;
        }
        input { 
            width: 100%; 
            padding: 12px 15px; 
            margin: 10px 0; 
            border: 1px solid #dcdde1; 
            border-radius: 8px; 
            font-family: inherit;
            font-size: 14px;
            background-color: #f8f9fa;
            outline: none;
        }
        input:focus {
            border-color: #3498db;
            background-color: #fff;
            box-shadow: 0 0 0 3px rgba(52, 152, 219, 0.15);
        }
        /* هماهنگ کردن ظاهر Placeholder با فونت جدید */
        input::placeholder {
            color: #aaa;
            font-size: 13px;
        }
        button { 
            width: 100%; 
            padding: 12px; 
            background: #2980b9; 
            color: white; 
            border: none; 
            border-radius: 8px; 
            cursor: pointer; 
            font-family: inherit;
            font-size: 16px; 
            font-weight: 600;
            margin-top: 15px;
            box-shadow: 0 4px 12px rgba(41, 128, 185, 0.2);
        }
        button:hover { 
            background: #2471a3; 
            transform: translateY(-1px);
            box-shadow: 0 6px 15px rgba(41, 128, 185, 0.3);
        }
        button:active {
            transform: translateY(1px);
        }
        .error { 
            color: #e74c3c; 
            background: #fadbd8;
            padding: 10px;
            border-radius: 8px;
            margin-bottom: 20px; 
            font-size: 13px; 
            border: 1px solid #f5b7b1;
        }
    </style>
</head>
<body>
    <div class="login-box">
        <h3>ورود پرسنل بندرعباس مال</h3>
        {{if .}} <div class="error">{{.}}</div> {{end}}
        <form method="POST" action="/login">
            <input type="text" name="username" placeholder="شماره تلفن همراه (نام کاربری)" required autocomplete="off">
            <input type="password" name="password" placeholder="کلمه عبور (همان شماره همراه)" required>
            <button type="submit">ورود به حساب</button>
        </form>
    </div>
</body>
</html>
`

// قالب HTML صفحه پروفایل و نمایش QR Code
const profileTemplate = `
<!DOCTYPE html>
<html lang="fa" dir="rtl">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>پروفایل پرسنل</title>
    <!-- اضافه کردن فونت زیبای وزیرمتن -->
    <link href="https://cdn.jsdelivr.net/gh/rastikerdar/vazirmatn@v33.003/Vazirmatn-font-face.css" rel="stylesheet" type="text/css" />
    <style>
        * {
            box-sizing: border-box;
            transition: all 0.3s ease;
        }
        body { 
            font-family: 'Vazirmatn', Tahoma, sans-serif; 
            background: linear-gradient(135deg, #e0eafc, #cfdef3); 
            padding: 20px; 
            margin: 0;
            display: flex;
            justify-content: center;
            align-items: center;
            min-height: 100vh;
        }
        .card { 
            background: rgba(255, 255, 255, 0.95); 
            width: 100%;
            max-width: 550px; 
            margin: 20px auto; 
            padding: 35px 30px; 
            border-radius: 16px; 
            box-shadow: 0 10px 30px rgba(0,0,0,0.06); 
        }
        h2 { 
            text-align: center;
            color: #2c3e50;
            margin-top: 0;
            margin-bottom: 30px;
            font-size: 22px;
            font-weight: 700;
            position: relative;
            padding-bottom: 12px;
        }
        /* خط تزیینی زیر عنوان */
        h2::after {
            content: '';
            position: absolute;
            bottom: 0;
            left: 50%;
            transform: translateX(-50%);
            width: 60px;
            height: 4px;
            background: #2980b9;
            border-radius: 2px;
        }
        .info-grid {
            display: grid;
            gap: 12px;
            margin-bottom: 30px;
        }
        .info-row {
            display: flex;
            justify-content: space-between;
            align-items: center;
            background: #f8f9fa;
            padding: 12px 16px;
            border-radius: 8px;
            border: 1px solid #edf2f7;
        }
        .info-row strong {
            color: #718096;
            font-size: 14px;
            font-weight: 600;
        }
        .info-row span {
            color: #2d3748;
            font-size: 15px;
            font-weight: 700;
        }
        .qr-section { 
            text-align: center; 
            margin-top: 25px; 
            border-top: 1px dashed #e2e8f0; 
            padding-top: 25px; 
        }
        .qr-section h4 {
            color: #4a5568;
            margin: 0 0 15px 0;
            font-size: 14px;
            font-weight: 600;
        }
        .qr-container {
            display: inline-block;
            background: white;
            padding: 12px;
            border-radius: 12px;
            border: 1px solid #e2e8f0;
            box-shadow: 0 4px 12px rgba(0,0,0,0.02);
        }
        .qr-section img { 
            width: 160px; 
            height: 160px; 
            display: block;
        }
        .logout-container {
            text-align: center;
            margin-top: 30px;
        }
        .logout { 
            display: inline-block; 
            padding: 10px 24px;
            color: #e53e3e; 
            background: #fff5f5;
            border: 1px solid #fed7d7;
            border-radius: 8px;
            text-decoration: none; 
            font-size: 14px;
            font-weight: 600;
        }
        .logout:hover {
            background: #e53e3e;
            color: white;
            border-color: #e53e3e;
            box-shadow: 0 4px 12px rgba(229, 62, 62, 0.2);
        }
    </style>
</head>
<body>
    <div class="card">
        <h2>اطلاعات پرسنلی</h2>
        
        <div class="info-grid">
            <div class="info-row">
                <strong>نام و نام خانوادگی:</strong>
                <span>{{.FullName}}</span>
            </div>
       
            <div class="info-row">
                <strong>تلفن همراه:</strong>
                <span>{{.MobilePhone}}</span>
            </div>
        </div>

        <div class="qr-section">
            <h4>کد QR اختصاصی شما</h4>
            <div class="qr-container">
                <img src="/qrcode" alt="QR Code">
            </div>
        </div>
        
        <div class="logout-container">
            <a href="/logout" class="logout">خروج از حساب کاربری</a>
        </div>
    </div>
</body>
</html>
`

// هندلر صفحه اصلی و پروفایل
func handleHome(w http.ResponseWriter, r *http.Request) {
	// چک کردن کوکی برای اصالت‌سنجی کاربر
	cookie, err := r.Cookie("session_user")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// پیدا کردن کاربر در اکسل
	user, found := findUser(cookie.Value)
	if !found {
		// خواسته شما: اگر کاربر از اکسل حذف شده بود خطای 404 بدهد
		http.Error(w, "404 یوزر یافت نشد", http.StatusNotFound)
		return
	}

	tmpl, _ := template.New("profile").Parse(profileTemplate)
	tmpl.Execute(w, user)
}

// هندلر صفحه لاگین
func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl, _ := template.New("login").Parse(loginTemplate)
		tmpl.Execute(w, nil)
		return
	}

	// پردازش اطلاعات فرم لاگین
	username := r.FormValue("username")
	password := r.FormValue("password")

	// بررسی صحت یوزرنیم و پسورد (طبق خواسته شما هر دو شماره همراه هستند)
	if username == password {
		_, found := findUser(username)
		if found {
			// ست کردن کوکی برای کاربر لود شده
			http.SetCookie(w, &http.Cookie{
				Name:    "session_user",
				Value:   username,
				Expires: time.Now().Add(24 * time.Hour),
				Path:    "/",
			})
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
	}

	// اگر اطلاعات غلط بود
	tmpl, _ := template.New("login").Parse(loginTemplate)
	tmpl.Execute(w, "شماره تلفن یا رمز عبور اشتباه است یا کاربر وجود ندارد.")
}

// هندلر تولید تصویر QR Code
func handleQRCode(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_user")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, found := findUser(cookie.Value)
	if !found {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// متن داخل QR Code طبق خواسته شما
	qrText := fmt.Sprintf("%s جزو پرسنل بندرعباس مال هست", user.FullName)

	// تولید QR کد به صورت آرایه‌ای از بایت‌ها (PNG)
	var png []byte
	png, err = qrcode.Encode(qrText, qrcode.Medium, 256)
	if err != nil {
		http.Error(w, "Error generating QR", http.StatusInternalServerError)
		return
	}

	// فرستادن تصویر به مرورگر
	w.Header().Set("Content-Type", "image/png")
	w.Write(png)
}

// هندلر خروج از حساب
func handleLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:    "session_user",
		Value:   "",
		Expires: time.Now().Add(-1 * time.Hour),
		Path:    "/",
	})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func main() {
	// لود اولیه فایل اکسل
	err := loadUsersFromExcel("users.xlsx")
	if err != nil {
		log.Printf("هشدار: فایل اکسل در ابتدا یافت نشد یا خطا دارد: %v", err)
	}

	// تعریف مسیرهای وب‌سایت (Routing)
	http.HandleFunc("/", handleHome)
	http.HandleFunc("/login", handleLogin)
	http.HandleFunc("/qrcode", handleQRCode)
	http.HandleFunc("/logout", handleLogout)

	fmt.Println("server run in the port 8080")
	fmt.Println("address for test :  http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

