// Copyright (C) 2023 Allen Li
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package codes contains return codes for the AniDB UDP API
package codes

// A ReturnCode is an AniDB UDP API return code.
// Note that even though ReturnCode implements error, not all
// ReturnCode values should be considered errors.
type ReturnCode int

const (
	LOGIN_ACCEPTED                    ReturnCode = 200
	LOGIN_ACCEPTED_NEW_VERSION        ReturnCode = 201
	LOGGED_OUT                        ReturnCode = 203
	RESOURCE                          ReturnCode = 205
	STATS                             ReturnCode = 206
	TOP                               ReturnCode = 207
	UPTIME                            ReturnCode = 208
	ENCRYPTION_ENABLED                ReturnCode = 209
	MYLIST_ENTRY_ADDED                ReturnCode = 210
	MYLIST_ENTRY_DELETED              ReturnCode = 211
	ADDED_FILE                        ReturnCode = 214
	ADDED_STREAM                      ReturnCode = 215
	EXPORT_QUEUED                     ReturnCode = 217
	EXPORT_CANCELLED                  ReturnCode = 218
	ENCODING_CHANGED                  ReturnCode = 219
	FILE                              ReturnCode = 220
	MYLIST                            ReturnCode = 221
	MYLIST_STATS                      ReturnCode = 222
	WISHLIST                          ReturnCode = 223
	NOTIFICATION                      ReturnCode = 224
	GROUP_STATUS                      ReturnCode = 225
	WISHLIST_ENTRY_ADDED              ReturnCode = 226
	WISHLIST_ENTRY_DELETED            ReturnCode = 227
	WISHLIST_ENTRY_UPDATED            ReturnCode = 228
	MULTIPLE_WISHLIST                 ReturnCode = 229
	ANIME                             ReturnCode = 230
	ANIME_BEST_MATCH                  ReturnCode = 231
	RANDOM_ANIME                      ReturnCode = 232
	ANIME_DESCRIPTION                 ReturnCode = 233
	REVIEW                            ReturnCode = 234
	CHARACTER                         ReturnCode = 235
	SONG                              ReturnCode = 236
	ANIMETAG                          ReturnCode = 237
	CHARACTERTAG                      ReturnCode = 238
	EPISODE                           ReturnCode = 240
	UPDATED                           ReturnCode = 243
	TITLE                             ReturnCode = 244
	CREATOR                           ReturnCode = 245
	NOTIFICATION_ENTRY_ADDED          ReturnCode = 246
	NOTIFICATION_ENTRY_DELETED        ReturnCode = 247
	NOTIFICATION_ENTRY_UPDATE         ReturnCode = 248
	MULTIPLE_NOTIFICATION             ReturnCode = 249
	GROUP                             ReturnCode = 250
	CATEGORY                          ReturnCode = 251
	BUDDY_LIST                        ReturnCode = 253
	BUDDY_STATE                       ReturnCode = 254
	BUDDY_ADDED                       ReturnCode = 255
	BUDDY_DELETED                     ReturnCode = 256
	BUDDY_ACCEPTED                    ReturnCode = 257
	BUDDY_DENIED                      ReturnCode = 258
	VOTED                             ReturnCode = 260
	VOTE_FOUND                        ReturnCode = 261
	VOTE_UPDATED                      ReturnCode = 262
	VOTE_REVOKED                      ReturnCode = 263
	HOT_ANIME                         ReturnCode = 265
	RANDOM_RECOMMENDATION             ReturnCode = 266
	RANDOM_SIMILAR                    ReturnCode = 267
	NOTIFICATION_ENABLED              ReturnCode = 270
	NOTIFYACK_SUCCESSFUL_MESSAGE      ReturnCode = 281
	NOTIFYACK_SUCCESSFUL_NOTIFICATION ReturnCode = 282
	NOTIFICATION_STATE                ReturnCode = 290
	NOTIFYLIST                        ReturnCode = 291
	NOTIFYGET_MESSAGE                 ReturnCode = 292
	NOTIFYGET_NOTIFY                  ReturnCode = 293
	SENDMESSAGE_SUCCESSFUL            ReturnCode = 294
	USER_ID                           ReturnCode = 295
	CALENDAR                          ReturnCode = 297

	PONG                                     ReturnCode = 300
	AUTHPONG                                 ReturnCode = 301
	NO_SUCH_RESOURCE                         ReturnCode = 305
	API_PASSWORD_NOT_DEFINED                 ReturnCode = 309
	FILE_ALREADY_IN_MYLIST                   ReturnCode = 310
	MYLIST_ENTRY_EDITED                      ReturnCode = 311
	MULTIPLE_MYLIST_ENTRIES                  ReturnCode = 312
	WATCHED                                  ReturnCode = 313
	SIZE_HASH_EXISTS                         ReturnCode = 314
	INVALID_DATA                             ReturnCode = 315
	STREAMNOID_USED                          ReturnCode = 316
	EXPORT_NO_SUCH_TEMPLATE                  ReturnCode = 317
	EXPORT_ALREADY_IN_QUEUE                  ReturnCode = 318
	EXPORT_NO_EXPORT_QUEUED_OR_IS_PROCESSING ReturnCode = 319
	NO_SUCH_FILE                             ReturnCode = 320
	NO_SUCH_ENTRY                            ReturnCode = 321
	MULTIPLE_FILES_FOUND                     ReturnCode = 322
	NO_SUCH_WISHLIST                         ReturnCode = 323
	NO_SUCH_NOTIFICATION                     ReturnCode = 324
	NO_GROUPS_FOUND                          ReturnCode = 325
	NO_SUCH_ANIME                            ReturnCode = 330
	NO_SUCH_DESCRIPTION                      ReturnCode = 333
	NO_SUCH_REVIEW                           ReturnCode = 334
	NO_SUCH_CHARACTER                        ReturnCode = 335
	NO_SUCH_SONG                             ReturnCode = 336
	NO_SUCH_ANIMETAG                         ReturnCode = 337
	NO_SUCH_CHARACTERTAG                     ReturnCode = 338
	NO_SUCH_EPISODE                          ReturnCode = 340
	NO_SUCH_UPDATES                          ReturnCode = 343
	NO_SUCH_TITLES                           ReturnCode = 344
	NO_SUCH_CREATOR                          ReturnCode = 345
	NO_SUCH_GROUP                            ReturnCode = 350
	NO_SUCH_CATEGORY                         ReturnCode = 351
	BUDDY_ALREADY_ADDED                      ReturnCode = 355
	NO_SUCH_BUDDY                            ReturnCode = 356
	BUDDY_ALREADY_ACCEPTED                   ReturnCode = 357
	BUDDY_ALREADY_DENIED                     ReturnCode = 358
	NO_SUCH_VOTE                             ReturnCode = 360
	INVALID_VOTE_TYPE                        ReturnCode = 361
	INVALID_VOTE_VALUE                       ReturnCode = 362
	PERMVOTE_NOT_ALLOWED                     ReturnCode = 363
	ALREADY_PERMVOTED                        ReturnCode = 364
	HOT_ANIME_EMPTY                          ReturnCode = 365
	RANDOM_RECOMMENDATION_EMPTY              ReturnCode = 366
	RANDOM_SIMILAR_EMPTY                     ReturnCode = 367
	NOTIFICATION_DISABLED                    ReturnCode = 370
	NO_SUCH_ENTRY_MESSAGE                    ReturnCode = 381
	NO_SUCH_ENTRY_NOTIFICATION               ReturnCode = 382
	NO_SUCH_MESSAGE                          ReturnCode = 392
	NO_SUCH_NOTIFY                           ReturnCode = 393
	NO_SUCH_USER                             ReturnCode = 394
	CALENDAR_EMPTY                           ReturnCode = 397
	NO_CHANGES                               ReturnCode = 399

	NOT_LOGGED_IN        ReturnCode = 403
	NO_SUCH_MYLIST_FILE  ReturnCode = 410
	NO_SUCH_MYLIST_ENTRY ReturnCode = 411
	MYLIST_UNAVAILABLE   ReturnCode = 412

	LOGIN_FAILED                   ReturnCode = 500
	LOGIN_FIRST                    ReturnCode = 501
	ACCESS_DENIED                  ReturnCode = 502
	CLIENT_VERSION_OUTDATED        ReturnCode = 503
	CLIENT_BANNED                  ReturnCode = 504
	ILLEGAL_INPUT_OR_ACCESS_DENIED ReturnCode = 505
	INVALID_SESSION                ReturnCode = 506
	NO_SUCH_ENCRYPTION_TYPE        ReturnCode = 509
	ENCODING_NOT_SUPPORTED         ReturnCode = 519
	BANNED                         ReturnCode = 555
	UNKNOWN_COMMAND                ReturnCode = 598

	INTERNAL_SERVER_ERROR ReturnCode = 600
	ANIDB_OUT_OF_SERVICE  ReturnCode = 601
	SERVER_BUSY           ReturnCode = 602
	NO_DATA               ReturnCode = 603
	TIMEOUT               ReturnCode = 604
	API_VIOLATION         ReturnCode = 666

	PUSHACK_CONFIRMED      ReturnCode = 701
	NO_SUCH_PACKET_PENDING ReturnCode = 702

	VERSION ReturnCode = 998
)

//go:generate stringer -type=ReturnCode -linecomment

func (c ReturnCode) Error() string {
	return c.String()
}
