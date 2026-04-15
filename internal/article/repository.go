package article

import (
	"database/sql"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

type feedRow struct {
	ArticleID     string         `db:"article_id"`
	Title         sql.NullString `db:"title"`
	Description   sql.NullString `db:"description"`
	Keywords      sql.NullString `db:"keywords"`
	URL           sql.NullString `db:"url"`
	Thumbnail     sql.NullString `db:"thumbnail"`
	LikeCount     int64          `db:"like_count"`
	BookmarkCount int64          `db:"bookmark_count"`
	ShareCount    int64          `db:"share_count"`
	PublishedAt   *time.Time     `db:"published_at"`
	CategoryID    *int64         `db:"category_id"`
	CategoryName  *string        `db:"category_name"`
	CategoryImage *string        `db:"category_image"`
	CategoryThumb *string        `db:"category_thumbnail"`
	BlogID        int64          `db:"blog_id"`
	BlogTitle     sql.NullString `db:"blog_title"`
	BlogFavicon   sql.NullString `db:"blog_favicon"`
	IsLiked       bool           `db:"is_liked"`
	IsArchived    bool           `db:"is_archived"`
	SortKey       sql.NullString `db:"sort_key"`
	HistoryID     *int64         `db:"history_id"`
}

func nullStr(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Like(memberID int64, articleID string) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	result, err := tx.Exec(`INSERT INTO article_like (article_id, member_id, is_deleted) VALUES (?, ?, false) ON DUPLICATE KEY UPDATE is_deleted = false, updated_at = IF(is_deleted, CURRENT_TIMESTAMP, updated_at)`, articleID, memberID)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	// affected=1: new insert, affected=2: is_deleted toggled to false
	if affected > 0 {
		if _, err := tx.Exec(`UPDATE article SET like_count = like_count + 1 WHERE article_id = ?`, articleID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *Repository) Unlike(memberID int64, articleID string) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	result, err := tx.Exec(`UPDATE article_like SET is_deleted = true, updated_at = CURRENT_TIMESTAMP WHERE article_id = ? AND member_id = ? AND is_deleted = false`, articleID, memberID)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected > 0 {
		if _, err := tx.Exec(`UPDATE article SET like_count = GREATEST(like_count - 1, 0) WHERE article_id = ?`, articleID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *Repository) Bookmark(memberID int64, articleID string) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	result, err := tx.Exec(`INSERT INTO article_bookmark (article_id, member_id, is_deleted) VALUES (?, ?, false) ON DUPLICATE KEY UPDATE is_deleted = false, updated_at = IF(is_deleted, CURRENT_TIMESTAMP, updated_at)`, articleID, memberID)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected > 0 {
		if _, err := tx.Exec(`UPDATE article SET bookmark_count = bookmark_count + 1 WHERE article_id = ?`, articleID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *Repository) Unbookmark(memberID int64, articleID string) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	result, err := tx.Exec(`UPDATE article_bookmark SET is_deleted = true, updated_at = CURRENT_TIMESTAMP WHERE article_id = ? AND member_id = ? AND is_deleted = false`, articleID, memberID)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected > 0 {
		if _, err := tx.Exec(`UPDATE article SET bookmark_count = GREATEST(bookmark_count - 1, 0) WHERE article_id = ?`, articleID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *Repository) Share(memberID int64, articleID string) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	result, err := tx.Exec(`INSERT IGNORE INTO article_share (article_id, member_id) VALUES (?, ?)`, articleID, memberID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows > 0 {
		if _, err := tx.Exec(`UPDATE article SET share_count = share_count + 1 WHERE article_id = ?`, articleID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *Repository) MarkUninterested(memberID int64, articleID string) error {
	_, err := r.db.Exec(`INSERT IGNORE INTO article_uninterest (article_id, member_id) VALUES (?, ?)`, articleID, memberID)
	return err
}

func (r *Repository) GetByIDs(memberID int64, articleIDs []string) ([]FeedItem, error) {
	return r.fetchFeedItems(memberID, `a.article_id IN (?)`, articleIDs, "FIELD(a.article_id, ?)")
}

func (r *Repository) ListBookmarks(memberID int64, limit int, from string) (*BookmarkPage, error) {
	condition := `ab.member_id = ? AND ab.is_deleted = false`
	args := []any{memberID}
	if from != "" {
		condition += ` AND a.sort_key < ?`
		args = append(args, from)
	}
	items, err := r.fetchRows(`
		SELECT a.article_id, a.title, a.description, a.keywords, a.url, a.thumbnail, a.like_count, a.bookmark_count, a.share_count, a.published_at,
		       a.category_id, i.name AS category_name, i.image AS category_image, i.thumbnail AS category_thumbnail,
		       b.blog_id, b.title AS blog_title, b.favicon AS blog_favicon,
		       EXISTS(SELECT 1 FROM article_like al WHERE al.article_id = a.article_id AND al.member_id = ? AND al.is_deleted = false) AS is_liked,
		       true AS is_archived,
		       a.sort_key
		FROM article_bookmark ab
		JOIN article a ON a.article_id = ab.article_id
		JOIN blog b ON b.blog_id = a.blog_id
		LEFT JOIN interest i ON i.interest_id = a.category_id
		WHERE `+condition+`
		ORDER BY ab.updated_at DESC
		LIMIT ?`, append([]any{memberID}, append(args, limit+1)...)...)
	if err != nil {
		return nil, err
	}
	page, next := paginateBySortKey(items, limit)
	return &BookmarkPage{Articles: page, Next: next}, nil
}

func (r *Repository) ListRecent(memberID int64, limit int, from string, lang string, blogID int64) (*RecentArticlesPage, error) {
	condition := `1 = 1`
	args := []any{memberID, memberID}
	if from != "" {
		condition += ` AND a.publish_sort_key < ?`
		args = append(args, from)
	}
	if blogID > 0 {
		condition += ` AND a.blog_id = ?`
		args = append(args, blogID)
	} else if lang == "ko" {
		condition += ` AND a.lang = 'ko'`
	} else if lang == "en" {
		condition += ` AND (a.lang != 'ko' OR a.lang IS NULL)`
	}
	items, err := r.fetchRows(`
		SELECT a.article_id, a.title, a.description, a.keywords, a.url, a.thumbnail, a.like_count, a.bookmark_count, a.share_count, a.published_at,
		       a.category_id, i.name AS category_name, i.image AS category_image, i.thumbnail AS category_thumbnail,
		       b.blog_id, b.title AS blog_title, b.favicon AS blog_favicon,
		       EXISTS(SELECT 1 FROM article_like al WHERE al.article_id = a.article_id AND al.member_id = ? AND al.is_deleted = false) AS is_liked,
		       EXISTS(SELECT 1 FROM article_bookmark ab WHERE ab.article_id = a.article_id AND ab.member_id = ? AND ab.is_deleted = false) AS is_archived,
		       a.publish_sort_key AS sort_key
		FROM article a
		JOIN blog b ON b.blog_id = a.blog_id
		LEFT JOIN interest i ON i.interest_id = a.category_id
		WHERE `+condition+`
		ORDER BY a.publish_sort_key DESC
		LIMIT ?`, append(args, limit+1)...)
	if err != nil {
		return nil, err
	}
	page, next := paginateBySortKey(items, limit)
	return &RecentArticlesPage{Articles: page, Next: next}, nil
}

func (r *Repository) ListByBlog(memberID int64, blogID int64, limit int, from string) (*RecentArticlesPage, error) {
	condition := `a.blog_id = ?`
	args := []any{memberID, memberID, blogID}
	if from != "" {
		condition += ` AND a.publish_sort_key < ?`
		args = append(args, from)
	}
	items, err := r.fetchRows(`
		SELECT a.article_id, a.title, a.description, a.keywords, a.url, a.thumbnail, a.like_count, a.bookmark_count, a.share_count, a.published_at,
		       a.category_id, i.name AS category_name, i.image AS category_image, i.thumbnail AS category_thumbnail,
		       b.blog_id, b.title AS blog_title, b.favicon AS blog_favicon,
		       EXISTS(SELECT 1 FROM article_like al WHERE al.article_id = a.article_id AND al.member_id = ? AND al.is_deleted = false) AS is_liked,
		       EXISTS(SELECT 1 FROM article_bookmark ab WHERE ab.article_id = a.article_id AND ab.member_id = ? AND ab.is_deleted = false) AS is_archived,
		       a.publish_sort_key AS sort_key
		FROM article a
		JOIN blog b ON b.blog_id = a.blog_id
		LEFT JOIN interest i ON i.interest_id = a.category_id
		WHERE `+condition+`
		ORDER BY a.publish_sort_key DESC
		LIMIT ?`, append(args, limit+1)...)
	if err != nil {
		return nil, err
	}
	page, next := paginateBySortKey(items, limit)
	return &RecentArticlesPage{Articles: page, Next: next}, nil
}

func (r *Repository) ListHistory(memberID int64, limit int, from *int64) (*ArticleHistoryPage, error) {
	condition := `ah.member_id = ?`
	args := []any{memberID, memberID, memberID}
	if from != nil {
		condition += ` AND ah.id < ?`
		args = append(args, *from)
	}
	items, err := r.fetchRows(`
		SELECT a.article_id, a.title, a.description, a.keywords, a.url, a.thumbnail, a.like_count, a.bookmark_count, a.share_count, a.published_at,
		       a.category_id, i.name AS category_name, i.image AS category_image, i.thumbnail AS category_thumbnail,
		       b.blog_id, b.title AS blog_title, b.favicon AS blog_favicon,
		       EXISTS(SELECT 1 FROM article_like al WHERE al.article_id = a.article_id AND al.member_id = ? AND al.is_deleted = false) AS is_liked,
		       EXISTS(SELECT 1 FROM article_bookmark ab WHERE ab.article_id = a.article_id AND ab.member_id = ? AND ab.is_deleted = false) AS is_archived,
		       a.sort_key,
		       ah.id AS history_id
		FROM article_history ah
		JOIN article a ON a.article_id = ah.article_id
		JOIN blog b ON b.blog_id = a.blog_id
		LEFT JOIN interest i ON i.interest_id = a.category_id
		WHERE `+condition+`
		ORDER BY ah.id DESC
		LIMIT ?`, append(args, limit+1)...)
	if err != nil {
		return nil, err
	}
	feedItems := make([]FeedItem, 0, min(len(items), limit))
	var next *int64
	for idx, row := range items {
		if idx >= limit {
			next = row.HistoryID
			break
		}
		feedItems = append(feedItems, row.toFeedItem())
	}
	return &ArticleHistoryPage{Articles: feedItems, Next: next}, nil
}

func (r *Repository) ListLikes(memberID int64, limit int, from string) (*LikedArticlesPage, error) {
	condition := `al.member_id = ? AND al.is_deleted = false`
	args := []any{memberID}
	if from != "" {
		condition += ` AND a.sort_key < ?`
		args = append(args, from)
	}
	items, err := r.fetchRows(`
		SELECT a.article_id, a.title, a.description, a.keywords, a.url, a.thumbnail, a.like_count, a.bookmark_count, a.share_count, a.published_at,
		       a.category_id, i.name AS category_name, i.image AS category_image, i.thumbnail AS category_thumbnail,
		       b.blog_id, b.title AS blog_title, b.favicon AS blog_favicon,
		       true AS is_liked,
		       EXISTS(SELECT 1 FROM article_bookmark ab WHERE ab.article_id = a.article_id AND ab.member_id = ? AND ab.is_deleted = false) AS is_archived,
		       a.sort_key
		FROM article_like al
		JOIN article a ON a.article_id = al.article_id
		JOIN blog b ON b.blog_id = a.blog_id
		LEFT JOIN interest i ON i.interest_id = a.category_id
		WHERE `+condition+`
		ORDER BY al.updated_at DESC
		LIMIT ?`, append([]any{memberID}, append(args, limit+1)...)...)
	if err != nil {
		return nil, err
	}
	page, next := paginateBySortKey(items, limit)
	return &LikedArticlesPage{Articles: page, Next: next}, nil
}

func (r *Repository) Search(query string, limit int, from string) (*ArticleSearchPage, error) {
	// The search endpoint is anonymous (no auth middleware), so there is
	// no member context to hydrate isLiked / isArchived. The web client
	// needs full article metadata to render each hit as a card, so we
	// run the same join as ListRecent with a hardcoded zero member_id —
	// both EXISTS sub-queries then return false for every row, which is
	// the correct "not interacted" state for an anonymous caller.
	likePattern := "%" + query + "%"
	condition := `a.title LIKE ? OR a.description LIKE ?`
	args := []any{int64(0), int64(0), likePattern, likePattern}
	if from != "" {
		condition = `(` + condition + `) AND a.publish_sort_key < ?`
		args = append(args, from)
	}
	items, err := r.fetchRows(`
		SELECT a.article_id, a.title, a.description, a.keywords, a.url, a.thumbnail, a.like_count, a.bookmark_count, a.share_count, a.published_at,
		       a.category_id, i.name AS category_name, i.image AS category_image, i.thumbnail AS category_thumbnail,
		       b.blog_id, b.title AS blog_title, b.favicon AS blog_favicon,
		       EXISTS(SELECT 1 FROM article_like al WHERE al.article_id = a.article_id AND al.member_id = ? AND al.is_deleted = false) AS is_liked,
		       EXISTS(SELECT 1 FROM article_bookmark ab WHERE ab.article_id = a.article_id AND ab.member_id = ? AND ab.is_deleted = false) AS is_archived,
		       a.publish_sort_key AS sort_key
		FROM article a
		JOIN blog b ON b.blog_id = a.blog_id
		LEFT JOIN interest i ON i.interest_id = a.category_id
		WHERE `+condition+`
		ORDER BY a.publish_sort_key DESC
		LIMIT ?`, append(args, limit+1)...)
	if err != nil {
		return nil, err
	}
	page, next := paginateBySortKey(items, limit)
	return &ArticleSearchPage{Articles: page, Next: next}, nil
}

func (r *Repository) GetArticleURL(articleID string) (string, error) {
	var url string
	err := r.db.Get(&url, `SELECT url FROM article WHERE article_id = ?`, articleID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	return url, nil
}

func (r *Repository) DeleteArticle(articleID string) error {
	_, err := r.db.Exec(`DELETE FROM article WHERE article_id = ?`, articleID)
	return err
}

func (r *Repository) fetchFeedItems(memberID int64, condition string, articleIDs []string, orderBy string) ([]FeedItem, error) {
	if len(articleIDs) == 0 {
		return []FeedItem{}, nil
	}
	query, args, err := sqlx.In(`
		SELECT a.article_id, a.title, a.description, a.keywords, a.url, a.thumbnail, a.like_count, a.bookmark_count, a.share_count, a.published_at,
		       a.category_id, i.name AS category_name, i.image AS category_image, i.thumbnail AS category_thumbnail,
		       b.blog_id, b.title AS blog_title, b.favicon AS blog_favicon,
		       EXISTS(SELECT 1 FROM article_like al WHERE al.article_id = a.article_id AND al.member_id = ? AND al.is_deleted = false) AS is_liked,
		       EXISTS(SELECT 1 FROM article_bookmark ab WHERE ab.article_id = a.article_id AND ab.member_id = ? AND ab.is_deleted = false) AS is_archived,
		       a.sort_key
		FROM article a
		JOIN blog b ON b.blog_id = a.blog_id
		LEFT JOIN interest i ON i.interest_id = a.category_id
		WHERE `+condition+`
		ORDER BY `+orderBy, memberID, memberID, articleIDs, articleIDs)
	if err != nil {
		return nil, err
	}
	query = r.db.Rebind(query)
	rows := make([]feedRow, 0)
	if err := r.db.Select(&rows, query, args...); err != nil {
		return nil, err
	}
	items := make([]FeedItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, row.toFeedItem())
	}
	return items, nil
}

func (r *Repository) fetchRows(query string, args ...any) ([]feedRow, error) {
	rows := make([]feedRow, 0)
	if err := r.db.Select(&rows, query, args...); err != nil {
		return nil, err
	}
	return rows, nil
}

func paginateBySortKey(rows []feedRow, limit int) ([]FeedItem, string) {
	items := make([]FeedItem, 0, min(len(rows), limit))
	var next string
	for idx, row := range rows {
		if idx >= limit {
			if row.SortKey.Valid {
				next = row.SortKey.String
			}
			break
		}
		items = append(items, row.toFeedItem())
	}
	return items, next
}

func (r feedRow) toFeedItem() FeedItem {
	kw := nullStr(r.Keywords)
	keywords := []string{}
	if strings.TrimSpace(kw) != "" {
		keywords = strings.Split(kw, "\t")
	}
	var category *Category
	if r.CategoryID != nil {
		category = &Category{ID: *r.CategoryID}
		if r.CategoryName != nil {
			category.Name = *r.CategoryName
		}
		if r.CategoryImage != nil {
			category.Image = *r.CategoryImage
		}
		if r.CategoryThumb != nil {
			category.Thumbnail = *r.CategoryThumb
		}
	}
	return FeedItem{
		ID:           r.ArticleID,
		Title:        strings.TrimSpace(nullStr(r.Title)),
		Description:  strings.TrimSpace(nullStr(r.Description)),
		Keywords:     keywords,
		URL:          nullStr(r.URL),
		Thumbnail:    nullStr(r.Thumbnail),
		LikeCount:    r.LikeCount,
		ArchiveCount: r.BookmarkCount,
		ShareCount:   r.ShareCount,
		PublishedAt:  r.PublishedAt,
		Category:     category,
		IsLiked:      r.IsLiked,
		IsArchived:   r.IsArchived,
		Blog: FeedBlogInfo{
			ID:      sqlIntToString(r.BlogID),
			Title:   nullStr(r.BlogTitle),
			Favicon: nullStr(r.BlogFavicon),
		},
	}
}

func sqlIntToString(value int64) string {
	return strconv.FormatInt(value, 10)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
