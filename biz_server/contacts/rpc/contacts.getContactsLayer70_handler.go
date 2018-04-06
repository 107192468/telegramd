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

package rpc

import (
	"github.com/golang/glog"
	"github.com/nebulaim/telegramd/baselib/logger"
	"github.com/nebulaim/telegramd/grpc_util"
	"github.com/nebulaim/telegramd/mtproto"
	"golang.org/x/net/context"
	"github.com/nebulaim/telegramd/biz/dal/dao"
	"github.com/nebulaim/telegramd/biz/core/user"
)

// @benqi: Android client
// contacts.getContactsLayer70#22c6aa08 hash:string = contacts.Contacts;
func (s *ContactsServiceImpl) ContactsGetContactsLayer70(ctx context.Context, request *mtproto.TLContactsGetContactsLayer70) (*mtproto.Contacts_Contacts, error) {
	md := grpc_util.RpcMetadataFromIncoming(ctx)
	glog.Infof("contacts.getContactsLayer70#22c6aa08 - metadata: %s, request: %s", logger.JsonDebugData(md), logger.JsonDebugData(request))

	contacts := mtproto.NewTLContactsContacts()

	contactsDOList := dao.GetUserContactsDAO(dao.DB_SLAVE).SelectUserContacts(md.UserId)
	if len(contactsDOList) == 0 {
		// contacts is nil
		contacts.SetSavedCount(0)
		return contacts.To_Contacts_Contacts(), nil
	}

	contacts.SetSavedCount(int32(len(contactsDOList)))

	userIdList := make([]int32, 0, len(contactsDOList))

	for _, do := range contactsDOList {
		contact := mtproto.NewTLContact()
		contact.SetUserId(do.ContactUserId)
		contact.SetMutual(mtproto.ToBool(true))
		contacts.Data2.Contacts = append(contacts.Data2.Contacts, contact.To_Contact())
		userIdList = append(userIdList, contact.GetUserId())
	}

	users := user.GetUserList(userIdList)
	for _, u := range users {
		if u.GetId() == md.UserId {
			u.SetSelf(true)
		} else {
			u.SetSelf(false)
		}
		u.SetContact(true)
		contacts.Data2.Users = append(contacts.Data2.Users, u.To_User())
	}
	// reply := mtproto.MakeContacts_Contacts(contacts)

	glog.Infof("contacts.getContactsLayer70#22c6aa08 - reply: %s\n", logger.JsonDebugData(contacts))
	return contacts.To_Contacts_Contacts(), nil
}
