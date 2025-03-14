package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"tech-blog/model"
	"tech-blog/repository"

	"github.com/labstack/echo/v4"
)

// ArticleCreateOutput ...
type ArticleCreateOutput struct {
	Article          *model.Article
	Message          string
	ValidationErrors []string
}

// ArticleIndex ...
func ArticleIndex(c echo.Context) error {
	// "/articles" のパスでリクエストがあったら "/" にリダイレクトします。
	// Google Analytics などでのアクセス解析時にパスが統一されて分析がしやすくなります。
	if c.Request().URL.Path == "/articles" {
		c.Redirect(http.StatusPermanentRedirect, "/")
	}

	// リポジトリの処理を呼び出して記事の一覧データを取得します。
	articles, err := repository.ArticleListByCursor(0)

	// エラーが発生した場合
	if err != nil {
		// エラー内容をサーバーのログに出力します。
		c.Logger().Error(err.Error())

		// クライアントにステータスコード 500 でレスポンスを返します。
		return c.NoContent(http.StatusInternalServerError)
	}

	// 取得できた最後の記事の ID をカーソルとして設定します。
	var cursor int
	if len(articles) != 0 {
		cursor = articles[len(articles)-1].ID
	}

	// テンプレートに渡すデータを map に格納します。
	data := map[string]interface{}{
		"Articles": articles,
		"Cursor":   cursor,
	}

	// テンプレートファイルとデータを指定して HTML を生成し、クライアントに返却します
	return render(c, "article/index.html", data)
}

// ArticleNew ...
func ArticleNew(c echo.Context) error {
	data := map[string]interface{}{
		"Message": "Article New",
		"Now":     time.Now(),
	}

	return render(c, "article/new.html", data)
}

// ArticleShow ...
func ArticleShow(c echo.Context) error {
	// パスパラメータから記事 ID を取得します。
	// 文字列型で取得されるので、strconv パッケージを利用して数値型にキャストしています。
	id, _ := strconv.Atoi(c.Param("articleID"))

	// 記事データを取得します。
	article, err := repository.ArticleGetByID(id)

	if err != nil {
		// エラー内容をサーバーのログに出力します。
		c.Logger().Error(err.Error())

		// ステータスコード 500 でレスポンスを返却します。
		return c.NoContent(http.StatusInternalServerError)
	}

	// テンプレートに渡すデータを map に格納します。
	data := map[string]interface{}{
		"Article": article,
	}

	// テンプレートファイルとデータを指定して HTML を生成し、クライアントに返却します。
	return render(c, "article/show.html", data)
}

// ArticleEdit ...
func ArticleEdit(c echo.Context) error {
	// パスパラメータから記事 ID を取得します。
	// 文字列型で取得されるので、strconv パッケージを利用して数値型にキャストしています。
	id, _ := strconv.Atoi(c.Param("articleID"))

	// 編集フォームの初期値として表示するために記事データを取得します。
	article, err := repository.ArticleGetByID(id)

	if err != nil {
		// エラー内容をサーバーのログに出力します。
		c.Logger().Error(err.Error())

		// ステータスコード 500 でレスポンスを返却します。
		return c.NoContent(http.StatusInternalServerError)
	}

	// テンプレートに渡すデータを map に格納します。
	data := map[string]interface{}{
		"Article": article,
	}

	// テンプレートファイルとデータを指定して HTML を生成し、クライアントに返却します。
	return render(c, "article/edit.html", data)
}

// ArticleCreate ...
func ArticleCreate(c echo.Context) error {
	// 送信されてくるフォームの内容を格納する構造体を宣言します。
	var article model.Article

	// レスポンスとして返却する構造体を宣言します。
	var out ArticleCreateOutput

	// フォームの内容を構造体に埋め込みます。
	if err := c.Bind(&article); err != nil {
		// エラーの内容をサーバーのログに出力します。
		c.Logger().Error(err.Error())

		// リクエストの解釈に失敗した場合は 400 エラーを返却します。
		return c.JSON(http.StatusBadRequest, out)
	}

	// バリデーションチェックを実行します。
	if err := c.Validate(&article); err != nil {
		// エラーの内容をサーバーのログに出力します。
		c.Logger().Error(err.Error())

		// エラー内容を検査してカスタムエラーメッセージを取得します。
		out.ValidationErrors = article.ValidationErrors(err)

		// 解釈できたパラメータが許可されていない値の場合は 422 エラーを返却します。
		return c.JSON(http.StatusUnprocessableEntity, out)
	}

	// repository を呼び出して保存処理を実行します。
	res, err := repository.ArticleCreate(&article)
	if err != nil {
		// エラーの内容をサーバーのログに出力します。
		c.Logger().Error(err.Error())

		// サーバー内の処理でエラーが発生した場合は 500 エラーを返却します。
		return c.JSON(http.StatusInternalServerError, out)
	}

	// SQL 実行結果から作成されたレコードの ID を取得します。
	id, _ := res.LastInsertId()

	// 構造体に ID をセットします。
	article.ID = int(id)

	// レスポンスの構造体に保存した記事のデータを格納します。
	out.Article = &article

	// 処理成功時はステータスコード 200 でレスポンスを返却します。
	return c.JSON(http.StatusOK, out)
}

// ArticleDelete ...
func ArticleDelete(c echo.Context) error {
	// パスパラメータから記事 ID を取得します。
	// 文字列型で取得されるので、strconv パッケージを利用して数値型にキャストしています。
	id, _ := strconv.Atoi(c.Param("articleID"))

	// repository の記事削除処理を呼び出します。
	if err := repository.ArticleDelete(id); err != nil {
		// サーバーのログにエラー内容を出力します。
		c.Logger().Error(err.Error())

		// サーバーサイドでエラーが発生した場合は 500 エラーを返却します。
		return c.JSON(http.StatusInternalServerError, "")
	}

	// 成功時はステータスコード 200 を返却します。
	return c.JSON(http.StatusOK, fmt.Sprintf("Article %d is deleted.", id))
}

// ArticleList ...
func ArticleList(c echo.Context) error {
	// クエリパラメータからカーソルの値を取得します。
	// 文字列型で取得できるので strconv パッケージを用いて数値型にキャストしています。
	cursor, _ := strconv.Atoi(c.QueryParam("cursor"))

	// リポジトリの処理を呼び出して記事の一覧データを取得します。
	// 引数にカーソルの値を渡して、ID のどの位置から 10 件取得するかを指定しています。
	articles, err := repository.ArticleListByCursor(cursor)

	// エラーが発生した場合
	if err != nil {
		// サーバーのログにエラー内容を出力します。
		c.Logger().Error(err.Error())

		// クライアントにステータスコード 500 でレスポンスを返します。
		// HTML ではなく JSON 形式でデータのみを返却するため、
		// c.HTMLBlob() ではなく c.JSON() を呼び出しています。
		return c.JSON(http.StatusInternalServerError, "")
	}

	// エラーがない場合は、ステータスコード 200 でレスポンスを返します。
	// JSON 形式で返却するため、c.HTMLBlob() ではなく c.JSON() を呼び出しています。
	return c.JSON(http.StatusOK, articles)
}

// ArticleUpdateOutput ...
type ArticleUpdateOutput struct {
	Article          *model.Article
	Message          string
	ValidationErrors []string
}

// ArticleUpdate ...
func ArticleUpdate(c echo.Context) error {
	// リクエスト送信元のパスを取得します。
	ref := c.Request().Referer()

	// リクエスト送信元のパスから記事 ID を抽出します。
	refID := strings.Split(ref, "/")[4]

	// リクエスト URL のパスパラメータから記事 ID を抽出します。
	reqID := c.Param("articleID")

	// 編集画面で表示している記事と更新しようとしている記事が異なる場合は、
	// 更新処理をせずに 400 エラーを返却します。
	if reqID != refID {
		return c.JSON(http.StatusBadRequest, "")
	}

	// フォームで送信される記事データを格納する構造体を宣言します。
	var article model.Article

	// レスポンスするデータの構造体を宣言します。
	var out ArticleUpdateOutput

	// フォームで送信されたデータを変数に格納します。
	if err := c.Bind(&article); err != nil {
		// リクエストのパラメータの解釈に失敗した場合は 400 エラーを返却します。
		return c.JSON(http.StatusBadRequest, out)
	}

	// 入力値のチェック（バリデーションチェック）を行います。
	if err := c.Validate(&article); err != nil {
		// エラー内容をレスポンスのフィールドに格納します。
		out.ValidationErrors = article.ValidationErrors(err)

		// 解釈できたパラメータが不正な値の場合は 422 エラーを返却します。
		return c.JSON(http.StatusUnprocessableEntity, out)
	}

	// 文字列型の ID を数値型にキャストします。
	articleID, _ := strconv.Atoi(reqID)

	// フォームデータを格納した構造体に ID をセットします。
	article.ID = articleID

	// 記事を更新する処理を呼び出します。
	_, err := repository.ArticleUpdate(&article)

	if err != nil {
		// レスポンスの構造体にエラー内容をセットします。
		out.Message = err.Error()

		// リクエスト自体は正しいにも関わらずサーバー側で処理が失敗した場合は 500 エラーを返却します。
		return c.JSON(http.StatusInternalServerError, out)
	}

	// レスポンスの構造体に記事データをセットします。
	out.Article = &article

	// 処理成功時はステータスコード 200 でレスポンスを返却します。
	return c.JSON(http.StatusOK, out)
}
