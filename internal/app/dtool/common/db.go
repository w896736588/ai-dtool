package common

import (
	"dev_tool/internal/app/dtool/define"
	"dev_tool/internal/pkg/p_common"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gsdb"
	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/spf13/cast"
)

var DbMain = &CSqlite{}

type CSqlite struct {
	Client *gsdb.GsSqlite
	Env    *define.Env
}

func (h *CSqlite) InitTable() {
	//TODO 初始化表机构和变更
}

func (h *CSqlite) Login(username, password string) (int, error) {
	one, err := h.Client.QuickQuery(`tbl_user`, `*`, map[string]interface{}{
		`username`: username,
		`password`: h.GetSaltPassword(password),
	}).One()
	if err != nil {
		return 0, err
	}
	if len(one) > 0 {
		return cast.ToInt(one[`id`]), nil
	} else {
		return 0, nil
	}
}

func (h *CSqlite) CreateUser(username, password string) (int, error) {
	one, err := h.Client.QuickQuery(`tbl_user`, `*`, map[string]interface{}{
		`username`: username,
	}).One()
	if err != nil {
		return 0, err
	}
	if len(one) > 0 {
		return 0, errors.New(`已存在用户`)
	}
	newId, newError := h.Client.QuickCreate(`tbl_user`, map[string]interface{}{
		`username`: username,
		`password`: h.GetSaltPassword(password),
	}).Exec()
	if newError != nil {
		return 0, newError
	}
	return cast.ToInt(newId), nil
}

func (h *CSqlite) GetSaltPassword(password string) string {
	return gstool.Md5(password + h.Env.AppName)
}

func (h *CSqlite) CreateSsh(name, host, port string, userid, isPublic int) (int, error) {
	newId, newError := h.Client.QuickCreate(`tbl_ssh`, map[string]interface{}{
		`name`:      name,
		`host`:      host,
		`port`:      port,
		`userid`:    userid,
		`is_public`: isPublic,
	}).Exec()
	if newError != nil {
		return 0, newError
	}
	return cast.ToInt(newId), nil
}

func (h *CSqlite) CreateSshUser(sshid, userid int, username, password string) (int, error) {
	one, err := h.Client.QuickQuery(`tbl_user_ssh`, `*`, map[string]interface{}{
		`ssh_id`:  sshid,
		`user_id`: userid,
	}).One()
	if err != nil {
		return 0, err
	}
	if len(one) > 0 {
		_, editError := h.Client.QuickUpdate(`tbl_user_ssh`, map[string]interface{}{
			`ssh_id`:  sshid,
			`user_id`: userid,
		}, map[string]interface{}{
			`username`: username,
			`password`: password,
		}).Exec()
		if editError != nil {
			return 0, editError
		}
		return cast.ToInt(one[`id`]), nil
	} else {
		newId, newError := h.Client.QuickCreate(`tbl_user_ssh`, map[string]interface{}{
			`ssh_id`:   sshid,
			`user_id`:  userid,
			`username`: username,
			`password`: password,
		}).Exec()
		if newError != nil {
			return 0, newError
		}
		return cast.ToInt(newId), nil
	}
}

func (h *CSqlite) CreateCode(name, path string, sshid, codeGroupId int) (int, error) {
	one, err := h.Client.QuickQuery(`tbl_user_ssh`, `*`, map[string]interface{}{
		`ssh_id`: sshid,
		`path`:   path,
	}).One()
	if err != nil {
		return 0, err
	}
	if len(one) > 0 {
		_, editError := h.Client.QuickUpdate(`tbl_user_ssh`, map[string]interface{}{
			`ssh_id`: sshid,
			`path`:   path,
		}, map[string]interface{}{
			`name`:          name,
			`code_group_id`: codeGroupId,
		}).Exec()
		if editError != nil {
			return 0, editError
		}
		return cast.ToInt(one[`id`]), nil
	} else {
		newId, newError := h.Client.QuickCreate(`tbl_user_ssh`, map[string]interface{}{
			`name`:          name,
			`path`:          path,
			`code_group_id`: codeGroupId,
			`ssh_id`:        sshid,
		}).Exec()
		if newError != nil {
			return 0, newError
		}
		return cast.ToInt(newId), nil
	}
}

func (h *CSqlite) GetSshConfigByUserSshId(userSshId int) (map[string]any, error) {
	sql := `select us.username,us.password,s.host,s.port from tbl_user_ssh us 
			left join tbl_ssh s on s.id = us.ssh_id 
			where us.id = ?`
	return h.Client.QueryBySql(sql, userSshId).One()
}

func (h *CSqlite) GetSshConfigByUserId(userId int) ([]map[string]any, error) {
	sql := `select us.username,us.password,s.host,s.port,us.id from tbl_user_ssh us 
			left join tbl_ssh s on s.id = us.ssh_id 
			where us.user_id = ?`
	return h.Client.QueryBySql(sql, userId).All()
}

func (h *CSqlite) GetAllSshConfig() ([]map[string]any, error) {
	return h.Client.QuickQuery(`tbl_ssh`, `*`, nil).All()
}

func (h *CSqlite) GetSshConfig(sshId any) (map[string]any, error) {
	return h.Client.QuickQuery(`tbl_ssh`, `*`, map[string]any{
		`id`: sshId,
	}).One()
}

func (h *CSqlite) GetRedisConfig(redisId any) (map[string]any, error) {
	return h.Client.QuickQuery(`tbl_redis`, `*`, map[string]any{
		`id`: redisId,
	}).One()
}

func (h *CSqlite) GetAllRedisConfig() ([]map[string]any, error) {
	return h.Client.QuickQuery(`tbl_redis`, `*`, nil).All()
}

func (h *CSqlite) GetMysqlConfig(mysqlId any) (map[string]any, error) {
	return h.Client.QuickQuery(`tbl_mysql`, `*`, map[string]any{
		`id`: mysqlId,
	}).One()
}

func (h *CSqlite) StarAdd(id, name, key, value, _type any) (int64, error) {
	if cast.ToInt(id) == 0 {
		return h.Client.QuickCreate(`tbl_star`, map[string]any{
			`name`:        name,
			`key`:         key,
			`value`:       value,
			`type`:        _type,
			`create_time`: time.Now().Unix(),
			`update_time`: time.Now().Unix(),
		}).Exec()
	} else {
		return h.Client.QuickUpdate(`tbl_star`, map[string]any{
			`id`: id,
		}, map[string]any{
			`name`:        name,
			`key`:         key,
			`value`:       value,
			`type`:        _type,
			`update_time`: time.Now().Unix(),
		}).Exec()
	}
}

func (h *CSqlite) StarDel(id any) (int64, error) {
	return h.Client.QuickDelete(`tbl_star`, map[string]any{
		`id`: id,
	}).Exec()
}

func (h *CSqlite) StarList(_type any) ([]map[string]any, error) {
	return h.Client.QuickQuery(`tbl_star`, `*`, map[string]any{
		`type`: _type,
	}).All()
}

func (h *CSqlite) MarkdownAdd(id, name, markdownType, value any) (int64, error) {
	if cast.ToInt(id) == 0 {
		return h.Client.QuickCreate(`tbl_markdown`, map[string]any{
			`name`:          name,
			`content`:       value,
			`markdown_type`: markdownType,
			`create_time`:   time.Now().Unix(),
			`update_time`:   time.Now().Unix(),
		}).Exec()
	} else {
		//记录变更记录
		oldInfo, _ := h.Client.QuickQuery(`tbl_markdown`, `content`, map[string]any{
			`id`: id,
		}).One()

		upNum, upErr := h.Client.QuickUpdate(`tbl_markdown`, map[string]any{
			`id`: id,
		}, map[string]any{
			`name`:        name,
			`content`:     value,
			`update_time`: time.Now().Unix(),
		}).Exec()
		if upErr == nil && upNum > 0 && cast.ToString(oldInfo[`content`]) != value {

			_, _ = h.Client.QuickCreate(`tbl_markdown_history`, map[string]any{
				`markdown_id`: id,
				`old_content`: oldInfo[`content`],
				`new_content`: value,
				`change_desc`: p_common.TBaseClient.DiffText(cast.ToString(oldInfo[`content`]), cast.ToString(value)),
				`create_time`: time.Now().Unix(),
				`update_time`: time.Now().Unix(),
			}).Exec()
		}
		return upNum, upErr
	}
}

func (h *CSqlite) MarkdownDel(id any) (int64, error) {
	return h.Client.QuickDelete(`tbl_markdown`, map[string]any{
		`id`: id,
	}).Exec()
}

func (h *CSqlite) MarkdownList(markdownType string) ([]map[string]any, error) {
	return h.Client.QuickQuery(`tbl_markdown`, `*`, map[string]any{
		`markdown_type`: markdownType,
	}).Order("weight asc").All()
}

func (h *CSqlite) MarkdownHistoryList(id int) ([]map[string]any, error) {
	return h.Client.QuickQuery(`tbl_markdown_history`, `*`, map[string]any{
		`markdown_id`: id,
	}).Order("id desc").All()
}

func (h *CSqlite) MarkdownHistoryDel(historyId any) (int64, error) {
	return h.Client.QuickDelete(`tbl_markdown_history`, map[string]any{
		`id`: historyId,
	}).Exec()
}

func (h *CSqlite) AllGlobal() ([]map[string]any, error) {
	return h.Client.QuickQuery(`tbl_global`, `*`, map[string]any{}).All()
}

func (h *CSqlite) AllGlobalMap() (map[string]any, error) {
	return h.Client.QuickQuery(`tbl_global`, `*`, map[string]any{}).ToMap(`key`, `value`)
}

func (h *CSqlite) CmdList(variableId any) ([]map[string]any, error) {
	return h.Client.QuickQuery(`tbl_variable_cmd`, `*`, map[string]any{
		`variable_id`: variableId,
		`status`:      1,
	}).Order(`weight asc`).All()
}

func (h *CSqlite) CmdInfo(cmdId any) (map[string]any, error) {
	return h.Client.QuickQuery(`tbl_variable_cmd`, `*`, map[string]any{
		`id`:     cmdId,
		`status`: 1,
	}).One()
}

func (h *CSqlite) Variable(variableId any) map[string]any {
	variableInfo, _ := h.Client.QuickQuery(`tbl_variable`, `*`, map[string]interface{}{
		`id`: variableId,
	}).One()
	return variableInfo
}

func (h *CSqlite) GetApiInfo(id int) (map[string]any, error) {
	return h.Client.QuickQuery(`tbl_api`, `*`, map[string]interface{}{
		`id`: id,
	}).One()
}

// MemoryFragmentList 查询记忆片段列表。
func (h *CSqlite) MemoryFragmentList(limit int) ([]map[string]any, error) {
	sql := `
select
	f.*,
	ifnull(group_concat(t.tag_name, ','), '') as tags_text
from tbl_memory_fragment f
left join tbl_memory_fragment_tag t on t.fragment_id = f.id
where f.is_deleted = 0
group by f.id
order by f.update_time desc, f.id desc`
	if limit > 0 {
		sql += fmt.Sprintf(" limit %d", limit)
	}
	list, err := h.Client.QueryBySql(sql).All()
	if err != nil {
		return nil, err
	}
	h.memoryFragmentFillDisplayFields(list)
	return list, nil
}

// MemoryFragmentInfo 查询单个记忆片段详情。
func (h *CSqlite) MemoryFragmentInfo(id int) (map[string]any, error) {
	one, err := h.Client.QuickQuery(`tbl_memory_fragment`, `*`, map[string]any{
		`id`:         id,
		`is_deleted`: 0,
	}).One()
	if err != nil {
		return nil, err
	}
	if len(one) == 0 {
		return nil, errors.New(`片段不存在`)
	}
	tags, tagErr := h.memoryFragmentLoadTags(id)
	if tagErr != nil {
		return nil, tagErr
	}
	one[`tags`] = tags
	one[`create_time_desc`] = h.memoryFragmentFormatTime(cast.ToInt64(one[`create_time`]))
	one[`update_time_desc`] = h.memoryFragmentFormatTime(cast.ToInt64(one[`update_time`]))
	one[`index_status_desc`] = h.memoryFragmentIndexStatusDesc(cast.ToString(one[`index_status`]))
	return one, nil
}

// MemoryFragmentSave 保存记忆片段。
func (h *CSqlite) MemoryFragmentSave(id int, title, content string, tags []string) (map[string]any, error) {
	now := time.Now().Unix()
	title = strings.TrimSpace(title)
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")
	if title == `` {
		title = `未命名片段`
	}
	tags = h.memoryFragmentNormalizeTags(tags)
	contentText := h.memoryFragmentToPlainText(content)
	indexStatus := `success`
	indexVersion := 1
	changeDesc := ``

	if id > 0 {
		oldInfo, err := h.Client.QuickQuery(`tbl_memory_fragment`, `*`, map[string]any{
			`id`:         id,
			`is_deleted`: 0,
		}).One()
		if err != nil {
			return nil, err
		}
		if len(oldInfo) == 0 {
			return nil, errors.New(`片段不存在`)
		}
		oldTags, tagErr := h.memoryFragmentLoadTags(id)
		if tagErr != nil {
			return nil, tagErr
		}
		indexVersion = cast.ToInt(oldInfo[`index_version`]) + 1
		_, err = h.Client.QuickUpdate(`tbl_memory_fragment`, map[string]any{
			`id`: id,
		}, map[string]any{
			`title`:         title,
			`content`:       content,
			`content_text`:  contentText,
			`index_status`:  `pending`,
			`index_version`: indexVersion,
			`update_time`:   now,
		}).Exec()
		if err != nil {
			return nil, err
		}
		if err = h.memoryFragmentReplaceTags(id, tags, now); err != nil {
			return nil, err
		}
		changeDesc = h.memoryFragmentBuildChangeDesc(
			cast.ToString(oldInfo[`title`]),
			title,
			cast.ToString(oldInfo[`content`]),
			content,
			oldTags,
			tags,
		)
		if changeDesc != `` {
			_, err = h.Client.QuickCreate(`tbl_memory_fragment_history`, map[string]any{
				`fragment_id`: id,
				`title_old`:   cast.ToString(oldInfo[`title`]),
				`title_new`:   title,
				`content_old`: cast.ToString(oldInfo[`content`]),
				`content_new`: content,
				`tags_old`:    gstool.JsonEncode(oldTags),
				`tags_new`:    gstool.JsonEncode(tags),
				`change_desc`: changeDesc,
				`create_time`: now,
				`update_time`: now,
			}).Exec()
			if err != nil {
				return nil, err
			}
		}
	} else {
		newID, err := h.Client.QuickCreate(`tbl_memory_fragment`, map[string]any{
			`title`:         title,
			`content`:       content,
			`content_text`:  contentText,
			`is_deleted`:    0,
			`index_status`:  `pending`,
			`index_version`: indexVersion,
			`create_time`:   now,
			`update_time`:   now,
		}).Exec()
		if err != nil {
			return nil, err
		}
		id = cast.ToInt(newID)
		if err = h.memoryFragmentReplaceTags(id, tags, now); err != nil {
			return nil, err
		}
	}

	if err := h.memoryFragmentSyncSearchIndex(id, title, contentText, tags, now); err != nil {
		indexStatus = `failed`
	} else {
		indexStatus = `success`
	}
	_, _ = h.Client.QuickUpdate(`tbl_memory_fragment`, map[string]any{
		`id`: id,
	}, map[string]any{
		`index_status`: indexStatus,
		`update_time`:  now,
	}).Exec()
	return h.MemoryFragmentInfo(id)
}

// MemoryFragmentSoftDelete 软删除记忆片段。
func (h *CSqlite) MemoryFragmentSoftDelete(id int) (int64, error) {
	now := time.Now().Unix()
	_, _ = h.Client.ExecBySql(`delete from tbl_memory_fragment_fts where fragment_id = ?`, id).Exec()
	_, _ = h.Client.ExecBySql(`delete from tbl_memory_fragment_vec where fragment_id = ?`, id).Exec()
	return h.Client.QuickUpdate(`tbl_memory_fragment`, map[string]any{
		`id`:         id,
		`is_deleted`: 0,
	}, map[string]any{
		`is_deleted`:   1,
		`index_status`: `pending`,
		`update_time`:  now,
	}).Exec()
}

// MemoryFragmentHistoryList 查询记忆片段历史记录。
func (h *CSqlite) MemoryFragmentHistoryList(fragmentID int) ([]map[string]any, error) {
	list, err := h.Client.QuickQuery(`tbl_memory_fragment_history`, `*`, map[string]any{
		`fragment_id`: fragmentID,
	}).Order("id desc").All()
	if err != nil {
		return nil, err
	}
	for i := range list {
		list[i][`create_time_desc`] = h.memoryFragmentFormatTime(cast.ToInt64(list[i][`create_time`]))
		tagsOld := make([]string, 0)
		tagsNew := make([]string, 0)
		_ = gstool.JsonDecode(cast.ToString(list[i][`tags_old`]), &tagsOld)
		_ = gstool.JsonDecode(cast.ToString(list[i][`tags_new`]), &tagsNew)
		list[i][`tags_old_list`] = tagsOld
		list[i][`tags_new_list`] = tagsNew
	}
	return list, nil
}

// MemoryFragmentTagList 查询记忆片段标签列表。
func (h *CSqlite) MemoryFragmentTagList() ([]map[string]any, error) {
	sql := `
select
	t.tag_name,
	count(1) as use_count
from tbl_memory_fragment_tag t
inner join tbl_memory_fragment f on f.id = t.fragment_id
where f.is_deleted = 0
group by t.tag_name
order by use_count desc, t.tag_name asc`
	return h.Client.QueryBySql(sql).All()
}

// MemoryFragmentSearch 搜索记忆片段。
func (h *CSqlite) MemoryFragmentSearch(mode, query string, selectedTags []string, limit int) ([]map[string]any, error) {
	mode = strings.TrimSpace(strings.ToLower(mode))
	if mode == `` {
		mode = `hybrid`
	}
	query = strings.TrimSpace(strings.ToLower(query))
	selectedTags = h.memoryFragmentNormalizeTags(selectedTags)
	sql := `
select
	f.*,
	ifnull(group_concat(distinct t.tag_name), '') as tags_text,
	ifnull(max(s.search_text), '') as search_text
from tbl_memory_fragment f
left join tbl_memory_fragment_tag t on t.fragment_id = f.id
left join tbl_memory_fragment_fts s on s.fragment_id = f.id
where f.is_deleted = 0
group by f.id`
	rows, err := h.Client.QueryBySql(sql).All()
	if err != nil {
		return nil, err
	}
	type searchRow struct {
		row   map[string]any
		score int
	}
	resultRows := make([]searchRow, 0)
	tokens := h.memoryFragmentSearchTokens(query)
	for _, row := range rows {
		tags := h.memoryFragmentSplitTagText(cast.ToString(row[`tags_text`]))
		if !h.memoryFragmentMatchSelectedTags(selectedTags, tags) {
			continue
		}
		searchText := strings.ToLower(cast.ToString(row[`search_text`]))
		match, score := h.memoryFragmentSearchScore(mode, query, tokens, searchText, tags, cast.ToString(row[`title`]))
		if query == `` && len(selectedTags) > 0 {
			match = true
			score = 1
		}
		if query == `` && len(selectedTags) == 0 {
			match = true
			score = 1
		}
		if !match {
			continue
		}
		rowCopy := row
		rowCopy[`tags`] = tags
		rowCopy[`create_time_desc`] = h.memoryFragmentFormatTime(cast.ToInt64(row[`create_time`]))
		rowCopy[`update_time_desc`] = h.memoryFragmentFormatTime(cast.ToInt64(row[`update_time`]))
		rowCopy[`index_status_desc`] = h.memoryFragmentIndexStatusDesc(cast.ToString(row[`index_status`]))
		rowCopy[`score`] = score
		resultRows = append(resultRows, searchRow{
			row:   rowCopy,
			score: score,
		})
	}
	sort.SliceStable(resultRows, func(i, j int) bool {
		if resultRows[i].score != resultRows[j].score {
			return resultRows[i].score > resultRows[j].score
		}
		updateLeft := cast.ToInt64(resultRows[i].row[`update_time`])
		updateRight := cast.ToInt64(resultRows[j].row[`update_time`])
		if updateLeft != updateRight {
			return updateLeft > updateRight
		}
		return cast.ToInt(resultRows[i].row[`id`]) > cast.ToInt(resultRows[j].row[`id`])
	})
	if limit <= 0 {
		limit = 50
	}
	result := make([]map[string]any, 0)
	for index, item := range resultRows {
		if index >= limit {
			break
		}
		result = append(result, item.row)
	}
	return result, nil
}

// memoryFragmentReplaceTags 重建片段标签。
func (h *CSqlite) memoryFragmentReplaceTags(fragmentID int, tags []string, now int64) error {
	if _, err := h.Client.ExecBySql(`delete from tbl_memory_fragment_tag where fragment_id = ?`, fragmentID).Exec(); err != nil {
		return err
	}
	for _, tag := range tags {
		if _, err := h.Client.QuickCreate(`tbl_memory_fragment_tag`, map[string]any{
			`fragment_id`: fragmentID,
			`tag_name`:    tag,
			`create_time`: now,
			`update_time`: now,
		}).Exec(); err != nil {
			return err
		}
	}
	return nil
}

// memoryFragmentLoadTags 查询片段标签。
func (h *CSqlite) memoryFragmentLoadTags(fragmentID int) ([]string, error) {
	list, err := h.Client.QuickQuery(`tbl_memory_fragment_tag`, `tag_name`, map[string]any{
		`fragment_id`: fragmentID,
	}).Order("tag_name asc").All()
	if err != nil {
		return nil, err
	}
	tags := make([]string, 0, len(list))
	for _, item := range list {
		tag := strings.TrimSpace(cast.ToString(item[`tag_name`]))
		if tag == `` {
			continue
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

// memoryFragmentSyncSearchIndex 同步记忆片段搜索索引。
func (h *CSqlite) memoryFragmentSyncSearchIndex(fragmentID int, title, contentText string, tags []string, now int64) error {
	tagText := strings.Join(tags, ` `)
	searchText := strings.TrimSpace(strings.ToLower(strings.Join([]string{title, contentText, tagText}, ` `)))
	if _, err := h.Client.ExecBySql(`delete from tbl_memory_fragment_fts where fragment_id = ?`, fragmentID).Exec(); err != nil {
		return err
	}
	if _, err := h.Client.ExecBySql(`delete from tbl_memory_fragment_vec where fragment_id = ?`, fragmentID).Exec(); err != nil {
		return err
	}
	if _, err := h.Client.QuickCreate(`tbl_memory_fragment_fts`, map[string]any{
		`fragment_id`:  fragmentID,
		`title`:        title,
		`content_text`: contentText,
		`tag_text`:     tagText,
		`search_text`:  searchText,
		`create_time`:  now,
		`update_time`:  now,
	}).Exec(); err != nil {
		return err
	}
	_, err := h.Client.QuickCreate(`tbl_memory_fragment_vec`, map[string]any{
		`fragment_id`:   fragmentID,
		`embedding`:     ``,
		`model_name`:    `pending`,
		`model_version`: `pending`,
		`create_time`:   now,
		`update_time`:   now,
	}).Exec()
	return err
}

// memoryFragmentBuildChangeDesc 生成历史摘要。
func (h *CSqlite) memoryFragmentBuildChangeDesc(oldTitle, newTitle, oldContent, newContent string, oldTags, newTags []string) string {
	changeParts := make([]string, 0)
	if oldTitle != newTitle {
		changeParts = append(changeParts, `标题`)
	}
	if oldContent != newContent {
		changeParts = append(changeParts, `内容`)
	}
	if !h.memoryFragmentStringSliceEqual(oldTags, newTags) {
		changeParts = append(changeParts, `标签`)
	}
	if len(changeParts) == 0 {
		return ``
	}
	return strings.Join(changeParts, `、`) + `已更新`
}

// memoryFragmentFillDisplayFields 填充列表展示字段。
func (h *CSqlite) memoryFragmentFillDisplayFields(list []map[string]any) {
	for i := range list {
		tags := h.memoryFragmentSplitTagText(cast.ToString(list[i][`tags_text`]))
		list[i][`tags`] = tags
		list[i][`create_time_desc`] = h.memoryFragmentFormatTime(cast.ToInt64(list[i][`create_time`]))
		list[i][`update_time_desc`] = h.memoryFragmentFormatTime(cast.ToInt64(list[i][`update_time`]))
		list[i][`index_status_desc`] = h.memoryFragmentIndexStatusDesc(cast.ToString(list[i][`index_status`]))
	}
}

// memoryFragmentFormatTime 格式化时间。
func (h *CSqlite) memoryFragmentFormatTime(unixTime int64) string {
	if unixTime <= 0 {
		return ``
	}
	return gstool.TimeUnixToString(time.Unix(unixTime, 0), `Y-m-d H:i:s`)
}

// memoryFragmentIndexStatusDesc 返回索引状态描述。
func (h *CSqlite) memoryFragmentIndexStatusDesc(status string) string {
	switch status {
	case `success`:
		return `索引成功`
	case `failed`:
		return `索引失败`
	default:
		return `待索引`
	}
}

// memoryFragmentNormalizeTags 规范化标签列表。
func (h *CSqlite) memoryFragmentNormalizeTags(tags []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0)
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		tag = strings.Trim(tag, ",，")
		if tag == `` {
			continue
		}
		lowerTag := strings.ToLower(tag)
		if seen[lowerTag] {
			continue
		}
		seen[lowerTag] = true
		result = append(result, tag)
	}
	sort.Strings(result)
	return result
}

// memoryFragmentSplitTagText 拆分标签文本。
func (h *CSqlite) memoryFragmentSplitTagText(tagText string) []string {
	if strings.TrimSpace(tagText) == `` {
		return []string{}
	}
	return h.memoryFragmentNormalizeTags(strings.Split(tagText, `,`))
}

// memoryFragmentToPlainText 提取 Markdown 纯文本。
func (h *CSqlite) memoryFragmentToPlainText(content string) string {
	text := strings.ReplaceAll(content, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	replacer := strings.NewReplacer(
		"```", " ",
		"`", " ",
		"#", " ",
		">", " ",
		"*", " ",
		"_", " ",
		"-", " ",
		"|", " ",
		"[", " ",
		"]", " ",
		"(", " ",
		")", " ",
	)
	text = replacer.Replace(text)
	linkReg := regexp.MustCompile(`https?://[^\s]+`)
	text = linkReg.ReplaceAllString(text, ` `)
	spaceReg := regexp.MustCompile(`\s+`)
	text = spaceReg.ReplaceAllString(text, ` `)
	return strings.TrimSpace(text)
}

// memoryFragmentStringSliceEqual 比较字符串切片是否一致。
func (h *CSqlite) memoryFragmentStringSliceEqual(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

// memoryFragmentSearchTokens 生成搜索词列表。
func (h *CSqlite) memoryFragmentSearchTokens(query string) []string {
	if query == `` {
		return []string{}
	}
	return strings.Fields(strings.ToLower(query))
}

// memoryFragmentMatchSelectedTags 判断标签筛选是否命中。
func (h *CSqlite) memoryFragmentMatchSelectedTags(selectedTags, rowTags []string) bool {
	if len(selectedTags) == 0 {
		return true
	}
	tagMap := make(map[string]bool)
	for _, tag := range rowTags {
		tagMap[strings.ToLower(tag)] = true
	}
	for _, tag := range selectedTags {
		if !tagMap[strings.ToLower(tag)] {
			return false
		}
	}
	return true
}

// memoryFragmentSearchScore 计算搜索得分。
func (h *CSqlite) memoryFragmentSearchScore(mode, query string, tokens []string, searchText string, tags []string, title string) (bool, int) {
	if query == `` {
		return true, 1
	}
	phraseMatched := strings.Contains(searchText, query)
	tokenMatchCount := 0
	allMatched := len(tokens) > 0
	for _, token := range tokens {
		if strings.Contains(searchText, token) {
			tokenMatchCount += 1
			continue
		}
		allMatched = false
	}
	tagScore := 0
	for _, tag := range tags {
		if strings.Contains(strings.ToLower(tag), query) {
			tagScore += 3
		}
	}
	titleScore := 0
	if strings.Contains(strings.ToLower(title), query) {
		titleScore = 6
	}
	switch mode {
	case `keyword`:
		if !allMatched && !phraseMatched {
			return false, 0
		}
		score := tokenMatchCount*10 + tagScore + titleScore
		if phraseMatched {
			score += 8
		}
		return true, score
	case `natural`:
		if tokenMatchCount == 0 && !phraseMatched {
			return false, 0
		}
		score := tokenMatchCount*6 + tagScore + titleScore
		if phraseMatched {
			score += 12
		}
		return true, score
	default:
		keywordMatch := allMatched || phraseMatched
		naturalMatch := tokenMatchCount > 0 || phraseMatched
		if !keywordMatch && !naturalMatch {
			return false, 0
		}
		score := tokenMatchCount*8 + tagScore + titleScore
		if keywordMatch {
			score += 8
		}
		if phraseMatched {
			score += 8
		}
		return true, score
	}
}
