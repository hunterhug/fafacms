package model

import (
	"fmt"
	"github.com/hunterhug/fafacms/core/util"
	"strings"
)

func CountContentAll(userId int64) (err error) {
	cns := make([]ContentNode, 0)
	err = FaFaRdb.Client.Where("user_id=?", userId).And("status=?", 0).Cols("id").Find(&cns)
	if err != nil {
		return
	}
	nodeIds := make([]string, 0, len(cns))
	for _, v := range cns {
		nodeIds = append(nodeIds, fmt.Sprintf("%d", v.Id))
	}

	inSql := ""
	if len(nodeIds) > 0 {
		inSql = strings.Join(nodeIds, ",")
		inSql = fmt.Sprintf("and node_id in (%s)", inSql)
	}

	// SELECT count(id) as count FROM `fafacms_content` WHERE first_publish_time!=0 and user_id=2 and version>0 and status!=1 and status!=3 and node_id in (1,2,3,4,5,6,7,8,9)
	sql := fmt.Sprintf("SELECT count(id) as count FROM `fafacms_content` WHERE first_publish_time!=0 and user_id=? and version>0 and status!=1 and status!=3 %s", inSql)
	fmt.Println(sql)
	result, err := FaFaRdb.Client.QueryString(sql, userId)
	if err != nil {
		return err
	}

	back := 0
	for _, v := range result {
		back, _ = util.SI(v["count"])
		break
	}

	u := new(User)
	u.ContentNum = int64(back)
	_, err = FaFaRdb.Client.Where("id=?", userId).Cols("content_num").Update(u)
	return
}

func CountContentOneNode(userId int64, nodeId int64) (err error) {
	sql := fmt.Sprintf("SELECT count(id) as count FROM `fafacms_content` WHERE first_publish_time!=0 and user_id=? and node_id=? and version>0 and status!=1 and status!=3")
	result, err := FaFaRdb.Client.QueryString(sql, userId, nodeId)
	if err != nil {
		return err
	}

	back := 0
	for _, v := range result {
		back, _ = util.SI(v["count"])
		break
	}

	n := new(ContentNode)
	n.ContentNum = int64(back)
	_, err = FaFaRdb.Client.Where("user_id=?", userId).Cols("content_num").Update(n)
	return
}

func CountContentCool(userId int64) (err error) {
	cns := make([]ContentNode, 0)
	err = FaFaRdb.Client.Where("user_id=?", userId).And("status=?", 0).Cols("id").Find(&cns)
	if err != nil {
		return
	}
	nodeIds := make([]string, 0, len(cns))
	for _, v := range cns {
		nodeIds = append(nodeIds, fmt.Sprintf("%d", v.Id))
	}

	inSql := ""
	if len(nodeIds) > 0 {
		inSql = strings.Join(nodeIds, ",")
		inSql = fmt.Sprintf("and node_id in (%s)", inSql)
	}

	sql := fmt.Sprintf("SELECT sum(cool) as count FROM `fafacms_content` WHERE first_publish_time!=0 and user_id=? and version>0 and status!=1 and status!=3 %s", inSql)
	result, err := FaFaRdb.Client.QueryString(sql, userId)
	if err != nil {
		return err
	}

	back := 0
	for _, v := range result {
		back, _ = util.SI(v["count"])
		break
	}

	u := new(User)
	u.ContentCoolNum = int64(back)
	_, err = FaFaRdb.Client.Where("id=?", userId).Cols("content_cool_num").Update(u)
	return
}
