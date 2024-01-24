package jira

type JiraIssueResult struct {
	Expand     string      `json:"expand"`
	StartAt    int         `json:"startAt"`
	MaxResults int         `json:"maxResults"`
	Total      int         `json:"total"`
	Issues     []JiraIssue `json:"issues"`
}

type JiraIssue struct {
	Expand string `json:"expand,omitempty"`
	ID     string `json:"id,omitempty"`
	Self   string `json:"self,omitempty"`
	Key    string `json:"key,omitempty"`
	Fields Fields `json:"fields,omitempty"`
}

type AvatarUrls struct {
	Four8X48  string `json:"48x48,omitempty"`
	Two4X24   string `json:"24x24,omitempty"`
	One6X16   string `json:"16x16,omitempty"`
	Three2X32 string `json:"32x32,omitempty"`
}

type Assignee struct {
	Self         string     `json:"self,omitempty"`
	AccountID    string     `json:"accountId,omitempty"`
	EmailAddress string     `json:"emailAddress,omitempty"`
	AvatarUrls   AvatarUrls `json:"avatarUrls,omitempty"`
	DisplayName  string     `json:"displayName,omitempty"`
	Active       bool       `json:"active,omitempty"`
	TimeZone     string     `json:"timeZone,omitempty"`
	AccountType  string     `json:"accountType,omitempty"`
}

type Reporter struct {
	Self         string     `json:"self,omitempty"`
	AccountID    string     `json:"accountId,omitempty"`
	EmailAddress string     `json:"emailAddress,omitempty"`
	AvatarUrls   AvatarUrls `json:"avatarUrls,omitempty"`
	DisplayName  string     `json:"displayName,omitempty"`
	Active       bool       `json:"active,omitempty"`
	TimeZone     string     `json:"timeZone,omitempty"`
	AccountType  string     `json:"accountType,omitempty"`
}

type Progress struct {
	Progress int `json:"progress,omitempty"`
	Total    int `json:"total,omitempty"`
}

type Votes struct {
	Self     string `json:"self,omitempty"`
	Votes    int    `json:"votes,omitempty"`
	HasVoted bool   `json:"hasVoted,omitempty"`
}

type Issuetype struct {
	Self           string `json:"self,omitempty"`
	ID             string `json:"id,omitempty"`
	Description    string `json:"description,omitempty"`
	IconURL        string `json:"iconUrl,omitempty"`
	Name           string `json:"name,omitempty"`
	Subtask        bool   `json:"subtask,omitempty"`
	AvatarID       int    `json:"avatarId,omitempty"`
	HierarchyLevel int    `json:"hierarchyLevel,omitempty"`
}

type ProjectCategory struct {
	Self        string `json:"self,omitempty"`
	ID          string `json:"id,omitempty"`
	Description string `json:"description,omitempty"`
	Name        string `json:"name,omitempty"`
}

type Project struct {
	Self            string          `json:"self,omitempty"`
	ID              string          `json:"id,omitempty"`
	Key             string          `json:"key,omitempty"`
	Name            string          `json:"name,omitempty"`
	ProjectTypeKey  string          `json:"projectTypeKey,omitempty"`
	Simplified      bool            `json:"simplified,omitempty"`
	AvatarUrls      AvatarUrls      `json:"avatarUrls,omitempty"`
	ProjectCategory ProjectCategory `json:"projectCategory,omitempty"`
}

type Watches struct {
	Self       string `json:"self,omitempty"`
	WatchCount int    `json:"watchCount,omitempty"`
	IsWatching bool   `json:"isWatching,omitempty"`
}

type Customfield10010 struct {
	Self  string `json:"self,omitempty"`
	Value string `json:"value,omitempty"`
	ID    string `json:"id,omitempty"`
}

type Customfield11221 struct {
	Self  string `json:"self,omitempty"`
	Value string `json:"value,omitempty"`
	ID    string `json:"id,omitempty"`
}

type Priority struct {
	Self    string `json:"self,omitempty"`
	IconURL string `json:"iconUrl,omitempty"`
	Name    string `json:"name,omitempty"`
	ID      string `json:"id,omitempty"`
}

type Customfield11429 struct {
	Self  string `json:"self,omitempty"`
	Value string `json:"value,omitempty"`
	ID    string `json:"id,omitempty"`
}

type StatusCategory struct {
	Self      string `json:"self,omitempty"`
	ID        int    `json:"id,omitempty"`
	Key       string `json:"key,omitempty"`
	ColorName string `json:"colorName,omitempty"`
	Name      string `json:"name,omitempty"`
}

type Status struct {
	Self           string         `json:"self,omitempty"`
	Description    string         `json:"description,omitempty"`
	IconURL        string         `json:"iconUrl,omitempty"`
	Name           string         `json:"name,omitempty"`
	ID             string         `json:"id,omitempty"`
	StatusCategory StatusCategory `json:"statusCategory,omitempty"`
}

type Creator struct {
	Self         string     `json:"self,omitempty"`
	AccountID    string     `json:"accountId,omitempty"`
	EmailAddress string     `json:"emailAddress,omitempty"`
	AvatarUrls   AvatarUrls `json:"avatarUrls,omitempty"`
	DisplayName  string     `json:"displayName,omitempty"`
	Active       bool       `json:"active,omitempty"`
	TimeZone     string     `json:"timeZone,omitempty"`
	AccountType  string     `json:"accountType,omitempty"`
}

type Aggregateprogress struct {
	Progress int `json:"progress,omitempty"`
	Total    int `json:"total,omitempty"`
}

type Customfield11181 struct {
	Self  string `json:"self,omitempty"`
	Value string `json:"value,omitempty"`
	ID    string `json:"id,omitempty"`
}

type Customfield11182 struct {
	Self  string `json:"self,omitempty"`
	Value string `json:"value,omitempty"`
	ID    string `json:"id,omitempty"`
}

type Customfield10400 struct {
	HasEpicLinkFieldDependency bool `json:"hasEpicLinkFieldDependency,omitempty"`
	ShowField                  bool `json:"showField,omitempty"`
}

type Fields struct {
	Customfield11160              any               `json:"customfield_11160,omitempty"`
	Customfield11282              any               `json:"customfield_11282,omitempty"`
	Customfield11161              any               `json:"customfield_11161,omitempty"`
	Customfield11162              any               `json:"customfield_11162,omitempty"`
	Customfield11163              any               `json:"customfield_11163,omitempty"`
	Customfield11284              any               `json:"customfield_11284,omitempty"`
	Customfield11164              any               `json:"customfield_11164,omitempty"`
	Customfield11287              any               `json:"customfield_11287,omitempty"`
	Customfield11166              any               `json:"customfield_11166,omitempty"`
	Customfield11167              any               `json:"customfield_11167,omitempty"`
	Resolution                    any               `json:"resolution,omitempty"`
	Customfield10500              string            `json:"customfield_10500,omitempty"`
	LastViewed                    any               `json:"lastViewed,omitempty"`
	Customfield11151              any               `json:"customfield_11151,omitempty"`
	Customfield11152              any               `json:"customfield_11152,omitempty"`
	Customfield11273              any               `json:"customfield_11273,omitempty"`
	Customfield11274              any               `json:"customfield_11274,omitempty"`
	Customfield11153              any               `json:"customfield_11153,omitempty"`
	Customfield11154              any               `json:"customfield_11154,omitempty"`
	Customfield11275              any               `json:"customfield_11275,omitempty"`
	Customfield11155              any               `json:"customfield_11155,omitempty"`
	Customfield11276              any               `json:"customfield_11276,omitempty"`
	Customfield11156              any               `json:"customfield_11156,omitempty"`
	Customfield11277              any               `json:"customfield_11277,omitempty"`
	Customfield11157              any               `json:"customfield_11157,omitempty"`
	Customfield11278              any               `json:"customfield_11278,omitempty"`
	Customfield11158              any               `json:"customfield_11158,omitempty"`
	Customfield11279              any               `json:"customfield_11279,omitempty"`
	Customfield11159              any               `json:"customfield_11159,omitempty"`
	Labels                        []any             `json:"labels,omitempty"`
	Aggregatetimeoriginalestimate any               `json:"aggregatetimeoriginalestimate,omitempty"`
	Issuelinks                    []any             `json:"issuelinks,omitempty"`
	Assignee                      Assignee          `json:"assignee,omitempty"`
	Components                    []any             `json:"components,omitempty"`
	Customfield11260              any               `json:"customfield_11260,omitempty"`
	Customfield11261              any               `json:"customfield_11261,omitempty"`
	Customfield11262              any               `json:"customfield_11262,omitempty"`
	Customfield11263              any               `json:"customfield_11263,omitempty"`
	Customfield11143              any               `json:"customfield_11143,omitempty"`
	Customfield11264              any               `json:"customfield_11264,omitempty"`
	Customfield11144              any               `json:"customfield_11144,omitempty"`
	Customfield11265              any               `json:"customfield_11265,omitempty"`
	Customfield11266              any               `json:"customfield_11266,omitempty"`
	Customfield11145              any               `json:"customfield_11145,omitempty"`
	Customfield11146              any               `json:"customfield_11146,omitempty"`
	Customfield11268              any               `json:"customfield_11268,omitempty"`
	Customfield11147              any               `json:"customfield_11147,omitempty"`
	Customfield11148              any               `json:"customfield_11148,omitempty"`
	Customfield11138              any               `json:"customfield_11138,omitempty"`
	Customfield11259              any               `json:"customfield_11259,omitempty"`
	Customfield11139              any               `json:"customfield_11139,omitempty"`
	Customfield10600              any               `json:"customfield_10600,omitempty"`
	Customfield11490              any               `json:"customfield_11490,omitempty"`
	Subtasks                      []any             `json:"subtasks,omitempty"`
	Customfield11250              any               `json:"customfield_11250,omitempty"`
	Customfield11251              any               `json:"customfield_11251,omitempty"`
	Customfield11252              any               `json:"customfield_11252,omitempty"`
	Reporter                      Reporter          `json:"reporter,omitempty"`
	Customfield11253              any               `json:"customfield_11253,omitempty"`
	Customfield11496              any               `json:"customfield_11496,omitempty"`
	Customfield11495              any               `json:"customfield_11495,omitempty"`
	Customfield11254              any               `json:"customfield_11254,omitempty"`
	Customfield11255              any               `json:"customfield_11255,omitempty"`
	Customfield11498              any               `json:"customfield_11498,omitempty"`
	Customfield11497              any               `json:"customfield_11497,omitempty"`
	Customfield11257              any               `json:"customfield_11257,omitempty"`
	Customfield11499              any               `json:"customfield_11499,omitempty"`
	Customfield11137              any               `json:"customfield_11137,omitempty"`
	Customfield11258              any               `json:"customfield_11258,omitempty"`
	Customfield11248              any               `json:"customfield_11248,omitempty"`
	Customfield11249              any               `json:"customfield_11249,omitempty"`
	Progress                      Progress          `json:"progress,omitempty"`
	Votes                         Votes             `json:"votes,omitempty"`
	Issuetype                     Issuetype         `json:"issuetype,omitempty"`
	Customfield11240              any               `json:"customfield_11240,omitempty"`
	Customfield11241              any               `json:"customfield_11241,omitempty"`
	Project                       Project           `json:"project,omitempty"`
	Customfield11242              any               `json:"customfield_11242,omitempty"`
	Customfield11000              any               `json:"customfield_11000,omitempty"`
	Customfield11485              any               `json:"customfield_11485,omitempty"`
	Customfield11243              any               `json:"customfield_11243,omitempty"`
	Customfield11123              any               `json:"customfield_11123,omitempty"`
	Customfield11244              any               `json:"customfield_11244,omitempty"`
	Customfield11245              any               `json:"customfield_11245,omitempty"`
	Customfield11246              any               `json:"customfield_11246,omitempty"`
	Customfield11489              any               `json:"customfield_11489,omitempty"`
	Customfield11247              any               `json:"customfield_11247,omitempty"`
	Customfield11488              any               `json:"customfield_11488,omitempty"`
	Customfield11237              any               `json:"customfield_11237,omitempty"`
	Customfield11238              any               `json:"customfield_11238,omitempty"`
	Customfield10700              any               `json:"customfield_10700,omitempty"`
	Customfield11239              any               `json:"customfield_11239,omitempty"`
	Resolutiondate                any               `json:"resolutiondate,omitempty"`
	Watches                       Watches           `json:"watches,omitempty"`
	Customfield11470              any               `json:"customfield_11470,omitempty"`
	Customfield11472              any               `json:"customfield_11472,omitempty"`
	Customfield11471              any               `json:"customfield_11471,omitempty"`
	Customfield11352              any               `json:"customfield_11352,omitempty"`
	Customfield11113              any               `json:"customfield_11113,omitempty"`
	Customfield11114              any               `json:"customfield_11114,omitempty"`
	Customfield11105              any               `json:"customfield_11105,omitempty"`
	Customfield11106              []any             `json:"customfield_11106,omitempty"`
	Customfield11107              any               `json:"customfield_11107,omitempty"`
	Customfield11108              string            `json:"customfield_11108,omitempty"`
	Customfield11109              any               `json:"customfield_11109,omitempty"`
	Updated                       string            `json:"updated,omitempty"`
	Timeoriginalestimate          any               `json:"timeoriginalestimate,omitempty"`
	Customfield11340              any               `json:"customfield_11340,omitempty"`
	Description                   any               `json:"description,omitempty"`
	Customfield10010              Customfield10010  `json:"customfield_10010,omitempty"`
	Customfield10011              string            `json:"customfield_10011,omitempty"`
	Customfield11221              Customfield11221  `json:"customfield_11221,omitempty"`
	Customfield11100              any               `json:"customfield_11100,omitempty"`
	Customfield11222              any               `json:"customfield_11222,omitempty"`
	Customfield11101              any               `json:"customfield_11101,omitempty"`
	Customfield10012              string            `json:"customfield_10012,omitempty"`
	Customfield11103              any               `json:"customfield_11103,omitempty"`
	Customfield11104              any               `json:"customfield_11104,omitempty"`
	Customfield11337              any               `json:"customfield_11337,omitempty"`
	Customfield10005              any               `json:"customfield_10005,omitempty"`
	Customfield10006              any               `json:"customfield_10006,omitempty"`
	Customfield11336              any               `json:"customfield_11336,omitempty"`
	Customfield11339              any               `json:"customfield_11339,omitempty"`
	Customfield10007              any               `json:"customfield_10007,omitempty"`
	Customfield11338              any               `json:"customfield_11338,omitempty"`
	Customfield10008              any               `json:"customfield_10008,omitempty"`
	Customfield10800              any               `json:"customfield_10800,omitempty"`
	Customfield10009              string            `json:"customfield_10009,omitempty"`
	Summary                       string            `json:"summary,omitempty"`
	Customfield11331              any               `json:"customfield_11331,omitempty"`
	Customfield10000              any               `json:"customfield_10000,omitempty"`
	Customfield11330              any               `json:"customfield_11330,omitempty"`
	Customfield11210              any               `json:"customfield_11210,omitempty"`
	Customfield11333              any               `json:"customfield_11333,omitempty"`
	Customfield10001              any               `json:"customfield_10001,omitempty"`
	Customfield11212              any               `json:"customfield_11212,omitempty"`
	Customfield11332              any               `json:"customfield_11332,omitempty"`
	Customfield11213              any               `json:"customfield_11213,omitempty"`
	Customfield10003              any               `json:"customfield_10003,omitempty"`
	Customfield11335              any               `json:"customfield_11335,omitempty"`
	Customfield11334              any               `json:"customfield_11334,omitempty"`
	Customfield11214              any               `json:"customfield_11214,omitempty"`
	Customfield10004              any               `json:"customfield_10004,omitempty"`
	Customfield11204              any               `json:"customfield_11204,omitempty"`
	Customfield11446              any               `json:"customfield_11446,omitempty"`
	Customfield11205              any               `json:"customfield_11205,omitempty"`
	Customfield11206              any               `json:"customfield_11206,omitempty"`
	Environment                   any               `json:"environment,omitempty"`
	Customfield11207              any               `json:"customfield_11207,omitempty"`
	Customfield11208              any               `json:"customfield_11208,omitempty"`
	Customfield11209              any               `json:"customfield_11209,omitempty"`
	Customfield11329              any               `json:"customfield_11329,omitempty"`
	Duedate                       any               `json:"duedate,omitempty"`
	Statuscategorychangedate      string            `json:"statuscategorychangedate,omitempty"`
	FixVersions                   []any             `json:"fixVersions,omitempty"`
	Customfield11200              any               `json:"customfield_11200,omitempty"`
	Customfield11201              []any             `json:"customfield_11201,omitempty"`
	Customfield11202              any               `json:"customfield_11202,omitempty"`
	Customfield11445              any               `json:"customfield_11445,omitempty"`
	Customfield11444              any               `json:"customfield_11444,omitempty"`
	Customfield11203              any               `json:"customfield_11203,omitempty"`
	Customfield10900              any               `json:"customfield_10900,omitempty"`
	Customfield10100              any               `json:"customfield_10100,omitempty"`
	Customfield10109              string            `json:"customfield_10109,omitempty"`
	Priority                      Priority          `json:"priority,omitempty"`
	Timeestimate                  any               `json:"timeestimate,omitempty"`
	Customfield11429              Customfield11429  `json:"customfield_11429,omitempty"`
	Versions                      []any             `json:"versions,omitempty"`
	Status                        Status            `json:"status,omitempty"`
	Customfield10203              any               `json:"customfield_10203,omitempty"`
	Aggregatetimeestimate         any               `json:"aggregatetimeestimate,omitempty"`
	Creator                       Creator           `json:"creator,omitempty"`
	Aggregateprogress             Aggregateprogress `json:"aggregateprogress,omitempty"`
	Customfield10200              any               `json:"customfield_10200,omitempty"`
	Customfield10201              any               `json:"customfield_10201,omitempty"`
	Customfield10202              any               `json:"customfield_10202,omitempty"`
	Customfield11405              any               `json:"customfield_11405,omitempty"`
	Customfield11407              any               `json:"customfield_11407,omitempty"`
	Customfield11406              any               `json:"customfield_11406,omitempty"`
	Timespent                     any               `json:"timespent,omitempty"`
	Aggregatetimespent            any               `json:"aggregatetimespent,omitempty"`
	Workratio                     int               `json:"workratio,omitempty"`
	Customfield11190              any               `json:"customfield_11190,omitempty"`
	Customfield11191              any               `json:"customfield_11191,omitempty"`
	Customfield11192              any               `json:"customfield_11192,omitempty"`
	Customfield11193              any               `json:"customfield_11193,omitempty"`
	Created                       string            `json:"created,omitempty"`
	Customfield11194              any               `json:"customfield_11194,omitempty"`
	Customfield11195              any               `json:"customfield_11195,omitempty"`
	Customfield11196              any               `json:"customfield_11196,omitempty"`
	Customfield11197              any               `json:"customfield_11197,omitempty"`
	Customfield11198              any               `json:"customfield_11198,omitempty"`
	Customfield11199              any               `json:"customfield_11199,omitempty"`
	Customfield11511              any               `json:"customfield_11511,omitempty"`
	Customfield10300              any               `json:"customfield_10300,omitempty"`
	Customfield11510              any               `json:"customfield_11510,omitempty"`
	Customfield11508              any               `json:"customfield_11508,omitempty"`
	Customfield11509              any               `json:"customfield_11509,omitempty"`
	Customfield11180              any               `json:"customfield_11180,omitempty"`
	Customfield11181              Customfield11181  `json:"customfield_11181,omitempty"`
	Customfield11182              Customfield11182  `json:"customfield_11182,omitempty"`
	Customfield11183              any               `json:"customfield_11183,omitempty"`
	Customfield11184              any               `json:"customfield_11184,omitempty"`
	Customfield11185              any               `json:"customfield_11185,omitempty"`
	Customfield11186              any               `json:"customfield_11186,omitempty"`
	Customfield11187              any               `json:"customfield_11187,omitempty"`
	Customfield11188              any               `json:"customfield_11188,omitempty"`
	Customfield11189              any               `json:"customfield_11189,omitempty"`
	Customfield11500              any               `json:"customfield_11500,omitempty"`
	Security                      any               `json:"security,omitempty"`
	Customfield11293              any               `json:"customfield_11293,omitempty"`
	Customfield11175              any               `json:"customfield_11175,omitempty"`
	Customfield11176              any               `json:"customfield_11176,omitempty"`
	Customfield11177              any               `json:"customfield_11177,omitempty"`
	Customfield11178              any               `json:"customfield_11178,omitempty"`
	Customfield11179              any               `json:"customfield_11179,omitempty"`
	Customfield10400              Customfield10400  `json:"customfield_10400,omitempty"`
}
