/*
 *  Copyright (c) 2017, https://github.com/nebulaim
 *  All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package dao

import (
	"github.com/nebulaim/telegramd/biz_model/dal/dao/mysql_dao"
	"github.com/jmoiron/sqlx"
	"github.com/golang/glog"
	"github.com/nebulaim/telegramd/baselib/redis_client"
	"github.com/nebulaim/telegramd/biz_model/dal/dao/redis_dao"
)

const (
	DB_MASTER 		= "immaster"
	DB_SLAVE 		= "imslave"
	CACHE 			= "cache"
)

type MysqlDAOList struct {
	// common
	CommonDAO *mysql_dao.CommonDAO

	// auth_key
	AuthKeysDAO  *mysql_dao.AuthKeysDAO
	AuthsDAO     *mysql_dao.AuthsDAO
	AuthSaltsDAO *mysql_dao.AuthSaltsDAO
	AuthUsersDAO *mysql_dao.AuthUsersDAO

	// biz
	UsersDAO                 *mysql_dao.UsersDAO
	DevicesDAO               *mysql_dao.DevicesDAO
	AuthPhoneTransactionsDAO *mysql_dao.AuthPhoneTransactionsDAO
	UserDialogsDAO           *mysql_dao.UserDialogsDAO
	UserContactsDAO          *mysql_dao.UserContactsDAO
	MessageBoxesDAO          *mysql_dao.MessageBoxesDAO
	MessagesDAO              *mysql_dao.MessagesDAO
	Messages2DAO             *mysql_dao.Messages2DAO
	UserNotifySettingsDAO    *mysql_dao.UserNotifySettingsDAO
	ReportsDAO               *mysql_dao.ReportsDAO
	UserPrivacysDAO          *mysql_dao.UserPrivacysDAO
	TmpPasswordsDAO          *mysql_dao.TmpPasswordsDAO
	ChatsDAO                 *mysql_dao.ChatsDAO
	ChatParticipantsDAO      *mysql_dao.ChatParticipantsDAO
	FilePartsDAO             *mysql_dao.FilePartsDAO
	FilesDAO                 *mysql_dao.FilesDAO
	PhotoDatasDAO            *mysql_dao.PhotoDatasDAO
	UserPtsUpdatesDAO        *mysql_dao.UserPtsUpdatesDAO
	UserQtsUpdatesDAO        *mysql_dao.UserQtsUpdatesDAO
	AuthSeqUpdatesDAO        *mysql_dao.AuthSeqUpdatesDAO
	AuthUpdatesStateDAO		 *mysql_dao.AuthUpdatesStateDAO
	UserPresencesDAO         *mysql_dao.UserPresencesDAO
}

// TODO(@benqi): 一主多从
type MysqlDAOManager struct {
	daoListMap map[string]*MysqlDAOList
}

var mysqlDAOManager = &MysqlDAOManager{make(map[string]*MysqlDAOList)}

func InstallMysqlDAOManager(clients map[string]*sqlx.DB) {
	for k, v := range clients {
		daoList := &MysqlDAOList{}

		// Common
		daoList.CommonDAO = mysql_dao.NewCommonDAO(v)

		// auth_key
		daoList.AuthKeysDAO = mysql_dao.NewAuthKeysDAO(v)
		daoList.AuthsDAO = mysql_dao.NewAuthsDAO(v)
		daoList.AuthSaltsDAO = mysql_dao.NewAuthSaltsDAO(v)
		daoList.AuthUsersDAO = mysql_dao.NewAuthUsersDAO(v)

		// biz
		daoList.UsersDAO = mysql_dao.NewUsersDAO(v)
		daoList.DevicesDAO = mysql_dao.NewDevicesDAO(v)
		daoList.AuthPhoneTransactionsDAO = mysql_dao.NewAuthPhoneTransactionsDAO(v)
		daoList.UserDialogsDAO = mysql_dao.NewUserDialogsDAO(v)
		daoList.UserContactsDAO = mysql_dao.NewUserContactsDAO(v)
		daoList.MessageBoxesDAO = mysql_dao.NewMessageBoxesDAO(v)
		daoList.MessagesDAO = mysql_dao.NewMessagesDAO(v)
		daoList.Messages2DAO = mysql_dao.NewMessages2DAO(v)
		daoList.AuthUpdatesStateDAO = mysql_dao.NewAuthUpdatesStateDAO(v)
		daoList.UserNotifySettingsDAO = mysql_dao.NewUserNotifySettingsDAO(v)
		daoList.ReportsDAO = mysql_dao.NewReportsDAO(v)
		daoList.UserPrivacysDAO = mysql_dao.NewUserPrivacysDAO(v)
		daoList.TmpPasswordsDAO = mysql_dao.NewTmpPasswordsDAO(v)
		daoList.ChatsDAO = mysql_dao.NewChatsDAO(v)
		daoList.ChatParticipantsDAO = mysql_dao.NewChatParticipantsDAO(v)
		daoList.FilePartsDAO = mysql_dao.NewFilePartsDAO(v)
		daoList.FilesDAO = mysql_dao.NewFilesDAO(v)
		daoList.PhotoDatasDAO = mysql_dao.NewPhotoDatasDAO(v)

		daoList.UserPtsUpdatesDAO = mysql_dao.NewUserPtsUpdatesDAO(v)
		daoList.UserQtsUpdatesDAO = mysql_dao.NewUserQtsUpdatesDAO(v)
		daoList.AuthSeqUpdatesDAO = mysql_dao.NewAuthSeqUpdatesDAO(v)

		daoList.UserPresencesDAO = mysql_dao.NewUserPresencesDAO(v)
		mysqlDAOManager.daoListMap[k] = daoList
	}
}

func  GetMysqlDAOListMap() map[string]*MysqlDAOList {
	return mysqlDAOManager.daoListMap
}

func  GetMysqlDAOList(dbName string) (daoList *MysqlDAOList) {
	daoList, ok := mysqlDAOManager.daoListMap[dbName]
	if !ok {
		glog.Errorf("GetMysqlDAOList - Not found daoList: %s", dbName)
	}
	return
}

func GetCommonDAO(dbName string) (dao *mysql_dao.CommonDAO) {
	daoList := GetMysqlDAOList(dbName)
	// err := mysqlDAOManager.daoListMap[dbName]
	if daoList != nil {
		dao = daoList.CommonDAO
	}
	return
}

func GetAuthKeysDAO(dbName string) (dao *mysql_dao.AuthKeysDAO) {
	daoList := GetMysqlDAOList(dbName)
	// err := mysqlDAOManager.daoListMap[dbName]
	if daoList != nil {
		dao = daoList.AuthKeysDAO
	}
	return
}

func GetAuthsDAO(dbName string) (dao *mysql_dao.AuthsDAO) {
	daoList := GetMysqlDAOList(dbName)
	// err := mysqlDAOManager.daoListMap[dbName]
	if daoList != nil {
		dao = daoList.AuthsDAO
	}
	return
}

func GetAuthSaltsDAO(dbName string) (dao *mysql_dao.AuthSaltsDAO) {
	daoList := GetMysqlDAOList(dbName)
	// err := mysqlDAOManager.daoListMap[dbName]
	if daoList != nil {
		dao = daoList.AuthSaltsDAO
	}
	return
}

func GetAuthUsersDAO(dbName string) (dao *mysql_dao.AuthUsersDAO) {
	daoList := GetMysqlDAOList(dbName)
	// err := mysqlDAOManager.daoListMap[dbName]
	if daoList != nil {
		dao = daoList.AuthUsersDAO
	}
	return
}

func GetUsersDAO(dbName string) (dao *mysql_dao.UsersDAO) {
	daoList := GetMysqlDAOList(dbName)
	// err := mysqlDAOManager.daoListMap[dbName]
	if daoList != nil {
		dao = daoList.UsersDAO
	}
	return
}

func GetDevicesDAO(dbName string) (dao *mysql_dao.DevicesDAO) {
	daoList := GetMysqlDAOList(dbName)
	// err := mysqlDAOManager.daoListMap[dbName]
	if daoList != nil {
		dao = daoList.DevicesDAO
	}
	return
}

func GetAuthPhoneTransactionsDAO(dbName string) (dao *mysql_dao.AuthPhoneTransactionsDAO) {
	daoList := GetMysqlDAOList(dbName)
	// err := mysqlDAOManager.daoListMap[dbName]
	if daoList != nil {
		dao = daoList.AuthPhoneTransactionsDAO
	}
	return
}

func GetUserDialogsDAO(dbName string) (dao *mysql_dao.UserDialogsDAO) {
	daoList := GetMysqlDAOList(dbName)
	// err := mysqlDAOManager.daoListMap[dbName]
	if daoList != nil {
		dao = daoList.UserDialogsDAO
	}
	return
}

func GetUserContactsDAO(dbName string) (dao *mysql_dao.UserContactsDAO) {
	daoList := GetMysqlDAOList(dbName)
	// err := mysqlDAOManager.daoListMap[dbName]
	if daoList != nil {
		dao = daoList.UserContactsDAO
	}
	return
}

func GetMessageBoxesDAO(dbName string) (dao *mysql_dao.MessageBoxesDAO) {
	daoList := GetMysqlDAOList(dbName)
	// err := mysqlDAOManager.daoListMap[dbName]
	if daoList != nil {
		dao = daoList.MessageBoxesDAO
	}
	return
}

func GetMessagesDAO(dbName string) (dao *mysql_dao.MessagesDAO) {
	daoList := GetMysqlDAOList(dbName)
	// err := mysqlDAOManager.daoListMap[dbName]
	if daoList != nil {
		dao = daoList.MessagesDAO
	}
	return
}

func GetMessages2DAO(dbName string) (dao *mysql_dao.Messages2DAO) {
	daoList := GetMysqlDAOList(dbName)
	// err := mysqlDAOManager.daoListMap[dbName]
	if daoList != nil {
		dao = daoList.Messages2DAO
	}
	return
}

func GetAuthUpdatesStateDAO(dbName string) (dao *mysql_dao.AuthUpdatesStateDAO) {
	daoList := GetMysqlDAOList(dbName)
	// err := mysqlDAOManager.daoListMap[dbName]
	if daoList != nil {
		dao = daoList.AuthUpdatesStateDAO
	}
	return
}

func GetUserNotifySettingsDAO(dbName string) (dao *mysql_dao.UserNotifySettingsDAO) {
	daoList := GetMysqlDAOList(dbName)
	// err := mysqlDAOManager.daoListMap[dbName]
	if daoList != nil {
		dao = daoList.UserNotifySettingsDAO
	}
	return
}

func GetReportsDAO(dbName string) (dao *mysql_dao.ReportsDAO) {
	daoList := GetMysqlDAOList(dbName)
	// err := mysqlDAOManager.daoListMap[dbName]
	if daoList != nil {
		dao = daoList.ReportsDAO
	}
	return
}

func GetUserPrivacysDAO(dbName string) (dao *mysql_dao.UserPrivacysDAO) {
	daoList := GetMysqlDAOList(dbName)
	// err := mysqlDAOManager.daoListMap[dbName]
	if daoList != nil {
		dao = daoList.UserPrivacysDAO
	}
	return
}

func GetTmpPasswordsDAO(dbName string) (dao *mysql_dao.TmpPasswordsDAO) {
	daoList := GetMysqlDAOList(dbName)
	// err := mysqlDAOManager.daoListMap[dbName]
	if daoList != nil {
		dao = daoList.TmpPasswordsDAO
	}
	return
}

func GetChatsDAO(dbName string) (dao *mysql_dao.ChatsDAO) {
	daoList := GetMysqlDAOList(dbName)
	// err := mysqlDAOManager.daoListMap[dbName]
	if daoList != nil {
		dao = daoList.ChatsDAO
	}
	return
}

func GetChatParticipantsDAO(dbName string) (dao *mysql_dao.ChatParticipantsDAO) {
	daoList := GetMysqlDAOList(dbName)
	// err := mysqlDAOManager.daoListMap[dbName]
	if daoList != nil {
		dao = daoList.ChatParticipantsDAO
	}
	return
}

func GetFilePartsDAO(dbName string) (dao *mysql_dao.FilePartsDAO) {
	daoList := GetMysqlDAOList(dbName)
	// err := mysqlDAOManager.daoListMap[dbName]
	if daoList != nil {
		dao = daoList.FilePartsDAO
	}
	return
}

func GetFilesDAO(dbName string) (dao *mysql_dao.FilesDAO) {
	daoList := GetMysqlDAOList(dbName)
	// err := mysqlDAOManager.daoListMap[dbName]
	if daoList != nil {
		dao = daoList.FilesDAO
	}
	return
}

func GetPhotoDatasDAO(dbName string) (dao *mysql_dao.PhotoDatasDAO) {
	daoList := GetMysqlDAOList(dbName)
	// err := mysqlDAOManager.daoListMap[dbName]
	if daoList != nil {
		dao = daoList.PhotoDatasDAO
	}
	return
}

func GetUserPtsUpdatesDAO(dbName string) (dao *mysql_dao.UserPtsUpdatesDAO) {
	daoList := GetMysqlDAOList(dbName)
	// err := mysqlDAOManager.daoListMap[dbName]
	if daoList != nil {
		dao = daoList.UserPtsUpdatesDAO
	}
	return
}

func GetUserQtsUpdatesDAO(dbName string) (dao *mysql_dao.UserQtsUpdatesDAO) {
	daoList := GetMysqlDAOList(dbName)
	// err := mysqlDAOManager.daoListMap[dbName]
	if daoList != nil {
		dao = daoList.UserQtsUpdatesDAO
	}
	return
}

func GetAuthSeqUpdatesDAO(dbName string) (dao *mysql_dao.AuthSeqUpdatesDAO) {
	daoList := GetMysqlDAOList(dbName)
	// err := mysqlDAOManager.daoListMap[dbName]
	if daoList != nil {
		dao = daoList.AuthSeqUpdatesDAO
	}
	return
}

func GetUserPresencesDAO(dbName string) (dao *mysql_dao.UserPresencesDAO) {
	daoList := GetMysqlDAOList(dbName)
	// err := mysqlDAOManager.daoListMap[dbName]
	if daoList != nil {
		dao = daoList.UserPresencesDAO
	}
	return
}

///////////////////////////////////////////////////////////////////////////////////////////
type RedisDAOList struct {
	SequenceDAO *redis_dao.SequenceDAO
}

type RedisDAOManager struct {
	daoListMap map[string]*RedisDAOList
}

var redisDAOManager = &RedisDAOManager{make(map[string]*RedisDAOList)}

func InstallRedisDAOManager(clients map[string]*redis_client.RedisPool) {
	for k, v := range clients {
		daoList := &RedisDAOList{}
		daoList.SequenceDAO = redis_dao.NewSequenceDAO(v)
		redisDAOManager.daoListMap[k] = daoList
	}
}

func  GetRedisDAOList(redisName string) (daoList *RedisDAOList) {
	daoList, ok := redisDAOManager.daoListMap[redisName]
	if !ok {
		glog.Errorf("GetRedisDAOList - Not found daoList: %s", redisName)
	}
	return
}

func  GetRedisDAOListMap() map[string]*RedisDAOList {
	return redisDAOManager.daoListMap
}

func GetSequenceDAO(redisName string) (dao *redis_dao.SequenceDAO) {
	daoList := GetRedisDAOList(redisName)
	// err := mysqlDAOManager.daoListMap[dbName]
	if daoList != nil {
		dao = daoList.SequenceDAO
	}
	return
}
