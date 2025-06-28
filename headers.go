package zeno

const (
	// HeaderAccept specifies media types acceptable for the response.
	HeaderAccept = "Accept"

	// HeaderAcceptCharset specifies acceptable character sets.
	HeaderAcceptCharset = "Accept-Charset"

	// HeaderAcceptEncoding specifies acceptable content encodings (e.g., gzip, deflate).
	HeaderAcceptEncoding = "Accept-Encoding"

	// HeaderAcceptLanguage specifies preferred natural languages for the response.
	HeaderAcceptLanguage = "Accept-Language"

	// HeaderAcceptRanges indicates that the server supports range requests.
	HeaderAcceptRanges = "Accept-Ranges"

	// HeaderAccessControlAllowOrigin specifies the origins that are allowed to access the resource.
	HeaderAccessControlAllowOrigin = "Access-Control-Allow-Origin"

	// HeaderAccessControlAllowMethods specifies the HTTP methods allowed when accessing the resource.
	HeaderAccessControlAllowMethods = "Access-Control-Allow-Methods"

	// HeaderAccessControlAllowHeaders specifies which headers can be used during the actual request.
	HeaderAccessControlAllowHeaders = "Access-Control-Allow-Headers"

	// HeaderAccessControlExposeHeaders indicates which headers can be exposed to the client.
	HeaderAccessControlExposeHeaders = "Access-Control-Expose-Headers"

	// HeaderAccessControlMaxAge indicates how long the results of a preflight request can be cached.
	HeaderAccessControlMaxAge = "Access-Control-Max-Age"

	// HeaderAccessControlRequestMethod is used in preflight requests to specify the method being used.
	HeaderAccessControlRequestMethod = "Access-Control-Request-Method"

	// HeaderAccessControlRequestHeaders is used in preflight requests to indicate which HTTP headers will be used.
	HeaderAccessControlRequestHeaders = "Access-Control-Request-Headers"

	// HeaderAge indicates the age of the object in a cache.
	HeaderAge = "Age"

	// HeaderAllow lists the allowed methods for a resource.
	HeaderAllow = "Allow"

	// HeaderAuthorization contains credentials for authenticating the user agent.
	HeaderAuthorization = "Authorization"

	// HeaderCacheControl specifies directives for caching mechanisms.
	HeaderCacheControl = "Cache-Control"

	// HeaderConnection controls whether the network connection stays open after the transaction finishes.
	HeaderConnection = "Connection"

	// HeaderContentDisposition indicates if the content should be displayed inline or as an attachment.
	HeaderContentDisposition = "Content-Disposition"

	// HeaderContentEncoding specifies the encoding used on the data.
	HeaderContentEncoding = "Content-Encoding"

	// HeaderContentLanguage describes the natural language(s) of the intended audience.
	HeaderContentLanguage = "Content-Language"

	// HeaderContentLength indicates the size of the message body.
	HeaderContentLength = "Content-Length"

	// HeaderContentLocation indicates an alternate location for the returned data.
	HeaderContentLocation = "Content-Location"

	// HeaderContentRange specifies where in a full body message a partial message belongs.
	HeaderContentRange = "Content-Range"

	// HeaderContentType indicates the media type of the resource.
	HeaderContentType = "Content-Type"

	// HeaderCookie contains stored HTTP cookies previously sent by the server.
	HeaderCookie = "Cookie"

	// HeaderDate represents the date and time at which the message was originated.
	HeaderDate = "Date"

	// HeaderETag is a unique identifier for a specific version of a resource.
	HeaderETag = "ETag"

	// HeaderExpect indicates expectations that need to be fulfilled before sending the request body.
	HeaderExpect = "Expect"

	// HeaderExpires gives the date/time after which the response is considered stale.
	HeaderExpires = "Expires"

	// HeaderForwarded is used to identify the originating IP address of a client.
	HeaderForwarded = "Forwarded"

	// HeaderFrom indicates the email address of the user making the request.
	HeaderFrom = "From"

	// HeaderHost specifies the domain name of the server and optionally the port.
	HeaderHost = "Host"

	// HeaderIfMatch makes the request conditional on the resource matching a given ETag.
	HeaderIfMatch = "If-Match"

	// HeaderIfModifiedSince makes the request conditional: it will only be successful if the resource has been modified since the given date.
	HeaderIfModifiedSince = "If-Modified-Since"

	// HeaderIfNoneMatch makes the request conditional: it will only be successful if the resource does not match the given ETag(s).
	HeaderIfNoneMatch = "If-None-Match"

	// HeaderIfRange makes a range request conditional based on ETag or modification date.
	HeaderIfRange = "If-Range"

	// HeaderIfUnmodifiedSince makes the request conditional: it will only be successful if the resource has not been modified since the given date.
	HeaderIfUnmodifiedSince = "If-Unmodified-Since"

	// HeaderLastModified indicates the date and time the resource was last modified.
	HeaderLastModified = "Last-Modified"

	// HeaderLink provides links to related resources.
	HeaderLink = "Link"

	// HeaderLocation indicates the URL to redirect a page to.
	HeaderLocation = "Location"

	// HeaderMaxForwards is used with TRACE and OPTIONS to limit the number of times a message can be forwarded.
	HeaderMaxForwards = "Max-Forwards"

	// HeaderOrigin indicates where the request originated (used in CORS).
	HeaderOrigin = "Origin"

	// HeaderPragma includes implementation-specific directives that may apply to any recipient along the request/response chain.
	HeaderPragma = "Pragma"

	// HeaderProxyAuthenticate defines the authentication method that should be used to access a resource behind a proxy.
	HeaderProxyAuthenticate = "Proxy-Authenticate"

	// HeaderProxyAuthorization contains credentials to authenticate a user agent with a proxy server.
	HeaderProxyAuthorization = "Proxy-Authorization"

	// HeaderRange specifies the range of bytes a client is requesting.
	HeaderRange = "Range"

	// HeaderReferer indicates the address of the previous web page from which a link to the currently requested page was followed.
	HeaderReferer = "Referer"

	// HeaderRetryAfter indicates how long the user agent should wait before making a follow-up request.
	HeaderRetryAfter = "Retry-After"

	// HeaderServer contains information about the software used by the origin server.
	HeaderServer = "Server"

	// HeaderSetCookie sends cookies from the server to the user agent.
	HeaderSetCookie = "Set-Cookie"

	// HeaderTE indicates what transfer encodings the client is willing to accept.
	HeaderTE = "TE"

	// HeaderTrailer indicates which headers will be present in the trailer part of the message.
	HeaderTrailer = "Trailer"

	// HeaderTransferEncoding specifies the form of encoding used to safely transfer the entity to the user.
	HeaderTransferEncoding = "Transfer-Encoding"

	// HeaderUpgrade is used to switch protocols (e.g., from HTTP/1.1 to WebSockets).
	HeaderUpgrade = "Upgrade"

	// HeaderUserAgent contains information about the user agent originating the request.
	HeaderUserAgent = "User-Agent"

	// HeaderVary indicates which headers a cache should use to decide if a cached response is fresh.
	HeaderVary = "Vary"

	// HeaderVia shows intermediate protocols and recipients between the user agent and server.
	HeaderVia = "Via"

	// HeaderWarning carries additional information about the status or transformation of a message.
	HeaderWarning = "Warning"

	// HeaderWWWAuthenticate indicates the authentication method that should be used to access a resource.
	HeaderWWWAuthenticate = "WWW-Authenticate"

	// HeaderAIM is a Microsoft extension for incremental GETs.
	HeaderAIM = "A-IM"

	// HeaderAcceptDatetime is used to request responses based on modification date (used in some APIs).
	HeaderAcceptDatetime = "Accept-Datetime"

	// HeaderAccessControlAllowCredentials indicates whether the response to the request can be exposed when credentials are present.
	HeaderAccessControlAllowCredentials = "Access-Control-Allow-Credentials"

	// HeaderAccessControlRequestCredentials is a non-standard header for requesting with credentials.
	HeaderAccessControlRequestCredentials = "Access-Control-Request-Credentials"

	// HeaderContentSecurityPolicy defines security policies (e.g., scripts, styles, etc.).
	HeaderContentSecurityPolicy = "Content-Security-Policy"

	// HeaderContentMD5 is a base64-encoded 128-bit MD5 digest of the message body.
	HeaderContentMD5 = "Content-MD5"

	// HeaderDNT indicates the user’s tracking preference (Do Not Track).
	HeaderDNT = "DNT"

	// HeaderDigest provides a message digest for the resource.
	HeaderDigest = "Digest"

	// HeaderEarlyData indicates if early data (0-RTT) is accepted.
	HeaderEarlyData = "Early-Data"

	// HeaderExpectCT allows sites to opt in to reporting of Certificate Transparency violations.
	HeaderExpectCT = "Expect-CT"

	// HeaderForwardedFor indicates the original IP address of a client connecting through a proxy.
	HeaderForwardedFor = "X-Forwarded-For"

	// HeaderForwardedHost indicates the original host requested by the client in the Host HTTP request header.
	HeaderForwardedHost = "X-Forwarded-Host"

	// HeaderForwardedProto indicates the originating protocol (HTTP or HTTPS).
	HeaderForwardedProto = "X-Forwarded-Proto"

	// HeaderHTTP2Settings is used in HTTP/2 to carry connection-specific settings.
	HeaderHTTP2Settings = "HTTP2-Settings"

	// HeaderKeepAlive specifies parameters for a persistent connection.
	HeaderKeepAlive = "Keep-Alive"

	// HeaderNEL reports network errors to a reporting endpoint.
	HeaderNEL = "NEL"

	// HeaderOriginTrial is used by Chrome to enable experimental web platform features.
	HeaderOriginTrial = "Origin-Trial"

	// HeaderPermissionsPolicy allows a server to declare permissions for APIs and features.
	HeaderPermissionsPolicy = "Permissions-Policy"

	// HeaderProxyConnection is a non-standard header used by some proxies.
	HeaderProxyConnection = "Proxy-Connection"

	// HeaderPublicKeyPins was used for Public Key Pinning (now deprecated).
	HeaderPublicKeyPins = "Public-Key-Pins"

	// HeaderReferrerPolicy specifies the referrer information sent with requests.
	HeaderReferrerPolicy = "Referrer-Policy"

	// HeaderReportTo specifies where reports should be sent in Reporting API.
	HeaderReportTo = "Report-To"

	// HeaderRequestID is used for request tracing (custom header).
	HeaderRequestID = "X-Request-ID"

	// HeaderSaveData indicates user preference for reduced data usage.
	HeaderSaveData = "Save-Data"

	// HeaderSourceMap indicates where source maps are located for debugging.
	HeaderSourceMap = "SourceMap"

	// HeaderStrictTransportSecurity enforces HTTPS communication with the server.
	HeaderStrictTransportSecurity = "Strict-Transport-Security"

	// HeaderTimingAllowOrigin specifies origins that are allowed to see timing information.
	HeaderTimingAllowOrigin = "Timing-Allow-Origin"

	// HeaderUpgradeInsecureRequests indicates that the client prefers secure content.
	HeaderUpgradeInsecureRequests = "Upgrade-Insecure-Requests"

	// HeaderXContentTypeOptions prevents MIME-sniffing (e.g., nosniff).
	HeaderXContentTypeOptions = "X-Content-Type-Options"

	// HeaderXDNSPrefetchControl controls DNS prefetching.
	HeaderXDNSPrefetchControl = "X-DNS-Prefetch-Control"

	// HeaderXDownloadOptions prevents automatic file opening in older browsers.
	HeaderXDownloadOptions = "X-Download-Options"

	// HeaderXFrameOptions prevents clickjacking by controlling frame usage.
	HeaderXFrameOptions = "X-Frame-Options"

	// HeaderXPermittedCrossDomainPolicies controls Adobe Flash and Acrobat access.
	HeaderXPermittedCrossDomainPolicies = "X-Permitted-Cross-Domain-Policies"

	// HeaderXPoweredBy is used to indicate the technology used (e.g., Express, PHP).
	HeaderXPoweredBy = "X-Powered-By"

	// HeaderXRequestID is a duplicate of X-Request-ID, used in some tracing setups.
	HeaderXRequestID = "X-Request-ID"

	// HeaderXUACompatible specifies compatibility mode for IE/Edge.
	HeaderXUACompatible = "X-UA-Compatible"

	// HeaderXXSSProtection enables cross-site scripting (XSS) filters.
	HeaderXXSSProtection = "X-XSS-Protection"

	// HeaderXRequestedWith is a non-standard header used to identify AJAX (XHR) requests.
	// Commonly set to "XMLHttpRequest" by client-side libraries like jQuery.
	HeaderXRequestedWith = "X-Requested-With"
)
