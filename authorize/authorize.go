package authorize

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"authorization/credentials"

	apiCaptcha "github.com/2captcha/2captcha-go"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

// Do - запускает процесс авторизации на сайте
// и возвращает куки файлы авторизованного запроса
func Do(credentials *credentials.Credentials) (string, error) {
	client := apiCaptcha.NewClient(credentials.RuCaptchaApiKey)
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// Запускаем браузер
	err := chromedp.Run(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to start browser: %w", err)
	}
	log.Printf("Browser starts")

	// Загружаем страницу
	err = chromedp.Run(ctx, chromedp.Navigate(credentials.LoginURL))
	if err != nil {
		return "", fmt.Errorf("failed to load page(%s): %w\n", credentials.LoginURL, err)
	}
	log.Printf("Login page %s loaded\n", credentials.LoginURL)

	// Заполняем форму авторизации
	err = fillLoginForm(ctx, credentials.Login, credentials.Password)
	if err != nil {
		return "", fmt.Errorf("failed to fill login form: %w", err)
	}
	time.Sleep(1 * time.Second)
	log.Println("Login form filled")

	// Получаем sitekey капчи
	var sitekey string
	err = chromedp.Run(ctx, chromedp.Evaluate(getSitekeyScript, &sitekey))
	if err != nil {
		return "", fmt.Errorf("failed to get captcha sitekey: %w", err)
	}
	log.Println("Captcha sitekey received")

	// Решаем капчу сервисом RuCaptcha, получаем токен
	token, err := solveReCaptcha(client, credentials.LoginURL, sitekey)
	if err != nil {
		return "", fmt.Errorf("failed to solve reCaptcha on third side service: %w", err)
	}
	log.Println("Captcha solved on third side service, token received")

	// Вставляем токен решения капчи в виджет гугл капчи
	err = chromedp.Run(ctx, chromedp.SetJavascriptAttribute(`#g-recaptcha-response`, "innerText", token))
	if err != nil {
		return "", fmt.Errorf("failed to insert captcha resolve token: %w", err)
	}
	log.Println("Token inserted")

	// Кликам кнопку Вход
	err = chromedp.Run(ctx, chromedp.Click(`.button.button_wide.button_primary`))
	if err != nil {
		return "", fmt.Errorf("chromedp.Click - Submit Button: %w", err)
	}
	time.Sleep(1 * time.Minute)
	log.Println("Form submitted")

	var aliasText string
	var exist bool
	// Создаем новый контекст с таймаутом
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Ищем ссылку на страницу авторизованного пользователя
	// для подтверждения нашей авторизации
	if err = chromedp.Run(ctxWithTimeout, chromedp.AttributeValue(
		`a.menu-head`,
		"href",
		&aliasText,
		&exist,
		chromedp.ByQuery,
	)); err != nil {
		if err == context.DeadlineExceeded {
			return "", fmt.Errorf("authorization failed: the operation took too long - timeout")
		}
		return "", fmt.Errorf("authorization failed: %w", err)
	}
	if !exist {
		return "", fmt.Errorf("authorization failed: unable to find authorized user data")
	}

	var cookiesSlice []string
	// Перезагружаем страницу и собираем куки запроса
	err = chromedp.Run(
		ctx,
		chromedp.Reload(),
		chromedp.ActionFunc(func(ctx context.Context) error {
			cookies, cookieErr := network.GetCookies().Do(ctx)
			if cookieErr != nil {
				return cookieErr
			}
			for _, cookie := range cookies {
				cookiesSlice = append(
					cookiesSlice,
					fmt.Sprintf("%s=%s", cookie.Name, cookie.Value),
				)
			}
			return nil
		}),
	)
	if err != nil {
		return "", fmt.Errorf("failed to get cookies: %w", err)
	}

	// собираем куки в формат значения хедера Cookie
	return strings.Join(cookiesSlice, "; "), nil
}

// solveReCaptcha - функция отправляет капчу на решение в сервис https://rucaptcha.com/
func solveReCaptcha(client *apiCaptcha.Client, targetURL, dataSiteKey string) (string, error) {
	c := apiCaptcha.ReCaptcha{
		SiteKey:   dataSiteKey,
		Url:       targetURL,
		Invisible: true,
		Action:    "verify",
	}

	return client.Solve(c.ToRequest())
}

// fillLoginForm - заполняет поля формы авторизации,
func fillLoginForm(ctx context.Context, log, pass string) error {
	err := chromedp.Run(ctx, chromedp.SetValue(`input[id="email_field"]`, log))
	if err != nil {
		return fmt.Errorf("chromedp.SetValue-email: %w", err)
	}
	err = chromedp.Run(ctx, chromedp.SetValue(`input[id="password_field"]`, pass))
	if err != nil {
		return fmt.Errorf("chromedp.SetValue-password: %w", err)
	}

	return nil
}

// JS скрипт для поиска и извлечения sitekey капчи
const getSitekeyScript = `
    var scripts = document.getElementsByTagName('script');
    for (var i = 0; i < scripts.length; i++) {
        var scriptContent = scripts[i].innerHTML;
        if (scriptContent.includes("'sitekey' : '")) {
            var startIndex = scriptContent.indexOf("'sitekey' : '") + "'sitekey' : '".length;
            var endIndex = scriptContent.indexOf("'", startIndex);
            sitekey = scriptContent.substring(startIndex, endIndex);
            break;
        }
    }
    sitekey;
`
