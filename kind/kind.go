package kind

import (
	"sync"

	"realy.lol/ints"
)

// T - which will be externally referenced as kind.T is the event type in the
// nostr protocol, the use of the capital T signifying type, consistent with Go
// idiom, the Go standard library, and much, conformant, existing code.
type T struct {
	K uint16
}

func New[V uint16 | uint32 | int32 | no](k V) (ki *T) { return &T{uint16(k)} }

func (k *T) ToInt() no {
	if k == nil {
		return 0
	}
	return no(k.K)
}
func (k *T) ToU16() uint16 {
	if k == nil {
		return 0
	}
	return k.K
}
func (k *T) ToI32() int32 {
	if k == nil {
		return 0
	}
	return int32(k.K)
}
func (k *T) ToU64() uint64 {
	if k == nil {
		return 0
	}
	return uint64(k.K)
}
func (k *T) Name() st       { return GetString(k) }
func (k *T) Equal(k2 *T) bo { return *k == *k2 }

var Privileged = []*T{
	EncryptedDirectMessage,
	GiftWrap,
	GiftWrapWithKind4,
	ApplicationSpecificData,
}

func (k *T) IsPrivileged() (is bo) {
	for i := range Privileged {
		if k.Equal(Privileged[i]) {
			return true
		}
	}
	return
}

func IsPrivileged(k ...*T) (is bo) {
	for _, kk := range k {
		for _, priv := range Privileged {
			if kk.Equal(priv) {
				return true
			}
		}
	}
	return
}

func (k *T) Marshal(dst by) (b by) { return ints.New(k.ToU64()).Marshal(dst) }

func (k *T) Unmarshal(b by) (r by, err er) {
	n := ints.New(0)
	if r, err = n.Unmarshal(b); chk.T(err) {
		return
	}
	k.K = n.Uint16()
	return
}

// GetString returns a human readable identifier for a kind.T.
func GetString(t *T) string {
	if t == nil {
		return ""
	}
	MapMx.Lock()
	defer MapMx.Unlock()
	return Map[t.K]
}

// IsEphemeral returns true if the event kind is an ephemeral event. (not to be
// stored)
func (k *T) IsEphemeral() bo {
	return k.K >= EphemeralStart.K && k.K < EphemeralEnd.K
}

// IsReplaceable returns true if the event kind is a replaceable kind - that is,
// if the newest version is the one that is in force (eg follow lists, relay
// lists, etc.
func (k *T) IsReplaceable() bo {
	return k.K == ProfileMetadata.K || k.K == FollowList.K ||
		(k.K >= ReplaceableStart.K && k.K < ReplaceableEnd.K)
}

// IsParameterizedReplaceable is a kind of event that is one of a group of
// events that replaces based on matching criteria.
func (k *T) IsParameterizedReplaceable() bo {
	return k.K >= ParameterizedReplaceableStart.K &&
		k.K < ParameterizedReplaceableEnd.K
}

// Directory events are events that necessarily need to be readable by anyone in
// order to interact with users who have access to the relay, in order to
// facilitate other users to find and interact with users on an auth-required
// relay.
var Directory = []*T{
	ProfileMetadata,
	FollowList,
	EventDeletion,
	Reporting,
	RelayListMetadata,
	MuteList,
	DMRelaysList,
}

// IsDirectoryEvent returns whether an event kind is a Directory event, which
// should grant permission to read such events without requiring authentication.
func (k *T) IsDirectoryEvent() bo {
	for i := range Directory {
		if k.Equal(Directory[i]) {
			return true
		}
	}
	return false
}

var (
	// ProfileMetadata is an event type that stores user profile data, pet
	// names, bio, lightning address, etc.
	ProfileMetadata = &T{0}
	// SetMetadata is a synonym for ProfileMetadata.
	SetMetadata = &T{0}
	// TextNote is a standard short text note of plain text a la twitter
	TextNote = &T{1}
	// RecommendServer is an event type that...
	RecommendServer = &T{2}
	RecommendRelay  = &T{2}
	// FollowList an event containing a list of pubkeys of users that should be
	// shown as follows in a timeline.
	FollowList = &T{3}
	Follows    = &T{3}
	// EncryptedDirectMessage is an event type that...
	EncryptedDirectMessage = &T{4}
	// Deletion is an event type that...
	Deletion      = &T{5}
	EventDeletion = &T{5}
	// Repost is an event type that...
	Repost = &T{6}
	// Reaction is an event type that...
	Reaction = &T{7}
	// BadgeAward is an event type
	BadgeAward = &T{8}
	// Seal is an event that wraps a PrivateDirectMessage and is placed inside a
	// GiftWrap or GiftWrapWithKind4
	Seal = &T{13}
	// PrivateDirectMessage is a nip-17 direct message with a different
	// construction. It doesn't actually appear as an event a relay might receive
	// but only as the stringified content of a GiftWrap or GiftWrapWithKind4 inside
	// a
	PrivateDirectMessage = &T{14}
	// ReadReceipt is a type of event that marks a list of tagged events (e
	// tags) as being seen by the client, its distinctive feature is the
	// "expiration" tag which indicates a time after which the marking expires
	ReadReceipt = &T{15}
	// GenericRepost is an event type that...
	GenericRepost = &T{16}
	// ChannelCreation is an event type that...
	ChannelCreation = &T{40}
	// ChannelMetadata is an event type that...
	ChannelMetadata = &T{41}
	// ChannelMessage is an event type that...
	ChannelMessage = &T{42}
	// ChannelHideMessage is an event type that...
	ChannelHideMessage = &T{43}
	// ChannelMuteUser is an event type that...
	ChannelMuteUser = &T{44}
	// Bid is an event type that...
	Bid = &T{1021}
	// BidConfirmation is an event type that...
	BidConfirmation = &T{1022}
	// OpenTimestamps is an event type that...
	OpenTimestamps    = &T{1040}
	GiftWrap          = &T{1059}
	GiftWrapWithKind4 = &T{1060}
	// FileMetadata is an event type that...
	FileMetadata = &T{1063}
	// LiveChatMessage is an event type that...
	LiveChatMessage = &T{1311}
	// BitcoinBlock is an event type created for the Nostrocket
	BitcoinBlock = &T{1517}
	// LiveStream from zap.stream
	LiveStream = &T{1808}
	// ProblemTracker is an event type used by Nostrocket
	ProblemTracker = &T{1971}
	// MemoryHole is an event type contains a report about an event (usually
	// text note or other human readable)
	MemoryHole = &T{1984}
	Reporting  = &T{1984}
	// Label is an event type has L and l tags, namespace and type - NIP-32
	Label = &T{1985}
	// CommunityPostApproval is an event type that...
	CommunityPostApproval = &T{4550}
	JobRequestStart       = &T{5000}
	JobRequestEnd         = &T{5999}
	JobResultStart        = &T{6000}
	JobResultEnd          = &T{6999}
	JobFeedback           = &T{7000}
	ZapGoal               = &T{9041}
	// ZapRequest is an event type that...
	ZapRequest = &T{9734}
	// Zap is an event type that...
	Zap        = &T{9735}
	Highlights = &T{9882}
	// ReplaceableStart is an event type that...
	ReplaceableStart = &T{10000}
	// MuteList is an event type that...
	MuteList  = &T{10000}
	BlockList = &T{10000}
	// PinList is an event type that...
	PinList = &T{10001}
	// RelayListMetadata is an event type that...
	RelayListMetadata     = &T{10002}
	BookmarkList          = &T{10003}
	CommunitiesList       = &T{10004}
	PublicChatsList       = &T{10005}
	BlockedRelaysList     = &T{10006}
	SearchRelaysList      = &T{10007}
	InterestsList         = &T{10015}
	UserEmojiList         = &T{10030}
	DMRelaysList          = &T{10050}
	FileStorageServerList = &T{10096}
	// NWCWalletInfo is an event type that...
	NWCWalletInfo = &T{13194}
	WalletInfo    = NWCWalletInfo
	// ReplaceableEnd is an event type that...
	ReplaceableEnd = &T{20000}
	// EphemeralStart is an event type that...
	EphemeralStart  = &T{20000}
	LightningPubRPC = &T{21000}
	// ClientAuthentication is an event type that...
	ClientAuthentication = &T{22242}
	// NWCWalletRequest is an event type that...
	NWCWalletRequest = &T{23194}
	WalletRequest    = &T{23194}
	// NWCWalletResponse is an event type that...
	NWCWalletResponse  = &T{23195}
	WalletResponse     = NWCWalletResponse
	NWCNotification    = &T{23196}
	WalletNotification = NWCNotification
	// NostrConnect is an event type that...
	NostrConnect = &T{24133}
	HTTPAuth     = &T{27235}
	// EphemeralEnd is an event type that...
	EphemeralEnd = &T{30000}
	// ParameterizedReplaceableStart is an event type that...
	ParameterizedReplaceableStart = &T{30000}
	// CategorizedPeopleList is an event type that...
	CategorizedPeopleList = &T{30000}
	FollowSets            = &T{30000}
	// CategorizedBookmarksList is an event type that...
	CategorizedBookmarksList = &T{30001}
	GenericLists             = &T{30001}
	RelaySets                = &T{30002}
	BookmarkSets             = &T{30003}
	CurationSets             = &T{30004}
	// ProfileBadges is an event type that...
	ProfileBadges = &T{30008}
	// BadgeDefinition is an event type that...
	BadgeDefinition = &T{30009}
	InterestSets    = &T{30015}
	// StallDefinition creates or updates a stall
	StallDefinition = &T{30017}
	// ProductDefinition creates or updates a product
	ProductDefinition    = &T{30018}
	MarketplaceUIUX      = &T{30019}
	ProductSoldAsAuction = &T{30020}
	// Article is an event type that...
	Article              = &T{30023}
	LongFormContent      = &T{30023}
	DraftLongFormContent = &T{30024}
	EmojiSets            = &T{30030}
	// ApplicationSpecificData is an event type stores data about application
	// configuration, this, like DMs and giftwraps must be protected by user
	// auth.
	ApplicationSpecificData = &T{30078}
	LiveEvent               = &T{30311}
	UserStatuses            = &T{30315}
	ClassifiedListing       = &T{30402}
	DraftClassifiedListing  = &T{30403}
	DateBasedCalendarEvent  = &T{31922}
	TimeBasedCalendarEvent  = &T{31923}
	Calendar                = &T{31924}
	CalendarEventRSVP       = &T{31925}
	HandlerRecommendation   = &T{31989}
	HandlerInformation      = &T{31990}
	// WaveLakeTrack which has no spec and uses malformed tags
	WaveLakeTrack       = &T{32123}
	CommunityDefinition = &T{34550}
	ACLEvent            = &T{39998}
	// ParameterizedReplaceableEnd is an event type that...
	ParameterizedReplaceableEnd = &T{40000}
)

var MapMx sync.Mutex
var Map = map[uint16]string{
	ProfileMetadata.K:             "ProfileMetadata",
	TextNote.K:                    "TextNote",
	RecommendRelay.K:              "RecommendRelay",
	FollowList.K:                  "FollowList",
	EncryptedDirectMessage.K:      "EncryptedDirectMessage",
	EventDeletion.K:               "EventDeletion",
	Repost.K:                      "Repost",
	Reaction.K:                    "Reaction",
	BadgeAward.K:                  "BadgeAward",
	ReadReceipt.K:                 "ReadReceipt",
	GenericRepost.K:               "GenericRepost",
	ChannelCreation.K:             "ChannelCreation",
	ChannelMetadata.K:             "ChannelMetadata",
	ChannelMessage.K:              "ChannelMessage",
	ChannelHideMessage.K:          "ChannelHideMessage",
	ChannelMuteUser.K:             "ChannelMuteUser",
	Bid.K:                         "Bid",
	BidConfirmation.K:             "BidConfirmation",
	OpenTimestamps.K:              "OpenTimestamps",
	FileMetadata.K:                "FileMetadata",
	LiveChatMessage.K:             "LiveChatMessage",
	ProblemTracker.K:              "ProblemTracker",
	Reporting.K:                   "Reporting",
	Label.K:                       "Label",
	CommunityPostApproval.K:       "CommunityPostApproval",
	JobRequestStart.K:             "JobRequestStart",
	JobRequestEnd.K:               "JobRequestEnd",
	JobResultStart.K:              "JobResultStart",
	JobResultEnd.K:                "JobResultEnd",
	JobFeedback.K:                 "JobFeedback",
	ZapGoal.K:                     "ZapGoal",
	ZapRequest.K:                  "ZapRequest",
	Zap.K:                         "Zap",
	Highlights.K:                  "Highlights",
	BlockList.K:                   "BlockList",
	PinList.K:                     "PinList",
	RelayListMetadata.K:           "RelayListMetadata",
	BookmarkList.K:                "BookmarkList",
	CommunitiesList.K:             "CommunitiesList",
	PublicChatsList.K:             "PublicChatsList",
	BlockedRelaysList.K:           "BlockedRelaysList",
	SearchRelaysList.K:            "SearchRelaysList",
	InterestsList.K:               "InterestsList",
	UserEmojiList.K:               "UserEmojiList",
	FileStorageServerList.K:       "FileStorageServerList",
	NWCWalletInfo.K:               "NWCWalletInfo",
	LightningPubRPC.K:             "LightningPubRPC",
	ClientAuthentication.K:        "ClientAuthentication",
	WalletRequest.K:               "WalletRequest",
	WalletResponse.K:              "WalletResponse",
	WalletNotification.K:          "WalletNotification",
	NostrConnect.K:                "NostrConnect",
	HTTPAuth.K:                    "HTTPAuth",
	FollowSets.K:                  "FollowSets",
	GenericLists.K:                "GenericLists",
	RelaySets.K:                   "RelaySets",
	BookmarkSets.K:                "BookmarkSets",
	CurationSets.K:                "CurationSets",
	ProfileBadges.K:               "ProfileBadges",
	BadgeDefinition.K:             "BadgeDefinition",
	InterestSets.K:                "InterestSets",
	StallDefinition.K:             "StallDefinition",
	ProductDefinition.K:           "ProductDefinition",
	MarketplaceUIUX.K:             "MarketplaceUIUX",
	ProductSoldAsAuction.K:        "ProductSoldAsAuction",
	LongFormContent.K:             "LongFormContent",
	DraftLongFormContent.K:        "DraftLongFormContent",
	EmojiSets.K:                   "EmojiSets",
	ApplicationSpecificData.K:     "ApplicationSpecificData",
	ParameterizedReplaceableEnd.K: "ParameterizedReplaceableEnd",
	LiveEvent.K:                   "LiveEvent",
	UserStatuses.K:                "UserStatuses",
	ClassifiedListing.K:           "ClassifiedListing",
	DraftClassifiedListing.K:      "DraftClassifiedListing",
	DateBasedCalendarEvent.K:      "DateBasedCalendarEvent",
	TimeBasedCalendarEvent.K:      "TimeBasedCalendarEvent",
	Calendar.K:                    "Calendar",
	CalendarEventRSVP.K:           "CalendarEventRSVP",
	HandlerRecommendation.K:       "HandlerRecommendation",
	HandlerInformation.K:          "HandlerInformation",
	CommunityDefinition.K:         "CommunityDefinition",
}
