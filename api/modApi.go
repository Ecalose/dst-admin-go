package api

import (
	"dst-admin-go/config/database"
	"dst-admin-go/mod"
	"dst-admin-go/model"
	"dst-admin-go/utils/clusterUtils"
	"dst-admin-go/utils/dstConfigUtils"
	"dst-admin-go/utils/fileUtils"
	"dst-admin-go/vo"
	"encoding/json"
	"log"
	"net/http"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type ModApi struct {
}

func (m *ModApi) SearchModList(ctx *gin.Context) {

	//获取查询参数
	text := ctx.Query("text")
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(ctx.DefaultQuery("size", "10"))

	data, err := mod.SearchModList(text, page, size)
	if err != nil {
		log.Panicln("搜索mod失败", err)
	}

	ctx.JSON(http.StatusOK, vo.Response{
		Code: 200,
		Msg:  "success",
		Data: data,
	})
}

func (m *ModApi) GetModInfo(ctx *gin.Context) {

	moId := ctx.Param("modId")
	modinfo, err, status := mod.GetModInfo(moId, ctx)
	if err != nil {
		log.Panicln("模组下载失败", "status: ", status)
	}
	var mod_config map[string]interface{}
	json.Unmarshal([]byte(modinfo.ModConfig), &mod_config)
	mod := map[string]interface{}{
		"auth":          modinfo.Auth,
		"consumer_id":   modinfo.ConsumerAppid,
		"creator_appid": modinfo.CreatorAppid,
		"description":   modinfo.Description,
		"file_url":      modinfo.FileUrl,
		"modid":         modinfo.Modid,
		"img":           modinfo.Img,
		"last_time":     modinfo.LastTime,
		"name":          modinfo.Name,
		"v":             modinfo.V,
		"mod_config":    mod_config,
	}
	ctx.JSON(http.StatusOK, vo.Response{
		Code: 200,
		Msg:  "success",
		Data: mod,
	})
}

func (m *ModApi) GetMyModList(ctx *gin.Context) {

	var modInfos []model.ModInfo
	db := database.DB

	db.Find(&modInfos)

	var modDataList []map[string]interface{}
	for _, modinfo := range modInfos {
		var mod_config map[string]interface{}
		json.Unmarshal([]byte(modinfo.ModConfig), &mod_config)
		mod := map[string]interface{}{
			"auth":          modinfo.Auth,
			"consumer_id":   modinfo.ConsumerAppid,
			"creator_appid": modinfo.CreatorAppid,
			"description":   modinfo.Description,
			"file_url":      modinfo.FileUrl,
			"modid":         modinfo.Modid,
			"img":           modinfo.Img,
			"last_time":     modinfo.LastTime,
			"name":          modinfo.Name,
			"v":             modinfo.V,
			"mod_config":    mod_config,
		}
		modDataList = append(modDataList, mod)
	}

	ctx.JSON(http.StatusOK, vo.Response{
		Code: 200,
		Msg:  "success",
		Data: modDataList,
	})

}

func (m *ModApi) DeleteMod(ctx *gin.Context) {

	modId := ctx.Param("modId")
	db := database.DB
	db.Where("modid = ?", modId).Delete(&model.ModInfo{})

	cluster := clusterUtils.GetClusterFromGin(ctx)
	mod_download_path := cluster.ModDownloadPath
	if mod_download_path == "" {
		log.Panicln("请设置模组下载路径")
	}
	mod_path := path.Join(mod_download_path, "/steamapps/workshop/content/322330/", modId)

	fileUtils.DeleteDir(mod_path)

	ctx.JSON(http.StatusOK, vo.Response{
		Code: 200,
		Msg:  "success",
		Data: modId,
	})
}

func (m *ModApi) DeleteSetupWorkshop(ctx *gin.Context) {
	dstPath := dstConfigUtils.GetDstConfig().Force_install_dir
	modsPath := filepath.Join(dstPath, "mods")
	// 删除所有workshop-xxx mod

	directories, err := fileUtils.ListDirectories(modsPath)
	if err != nil {
		log.Panicln("delete dst workshop file error", err)
	}
	var workshopList []string
	for _, directory := range directories {
		if strings.Contains(directory, "workshop") {
			workshopList = append(workshopList, directory)
		}
	}
	for _, workshop := range workshopList {
		err := fileUtils.DeleteDir(workshop)
		if err != nil {
			return
		}
	}

	ctx.JSON(http.StatusOK, vo.Response{
		Code: 200,
		Msg:  "success",
		Data: nil,
	})
}
