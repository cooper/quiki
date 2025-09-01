# quiki architecture overview

**quiki** go-based wiki system with custom markup parser (wikifier), web admin interface (adminifier), and sophisticated content management.

## core components

### wikifier package - parsing engine
transforms wiki markup into html via block-based parsing

**block types:**
- `sec{}` - hierarchical sections/headings
- `p{}` - paragraph text content  
- `infobox{}` - structured key-value boxes
- `gallery{}` - image galleries (nanogallery2)
- `image{}`/`imagebox{}` - images with server-side sizing
- `list{}` - ordered/unordered lists
- `code{}` - syntax highlighted code
- `map{}` - key-value structures (base for other blocks)
- `for{}` - dynamic loop iteration
- `html{}` - raw html embedding
- `style{}` - css styling

**parsing pipeline:**
1. **tokenization** - character-by-character parsing
2. **block recognition** - identifies types, creates objects
3. **content parsing** - block-specific content handling
4. **variable resolution** - processes variables/conditionals
5. **html generation** - converts to html output

**key files:**
- `parser.go` - master parser with position tracking
- `block-manager.go` - block registration/creation
- `block-*.go` - individual block implementations
- `variable.go` - variable system with scoping

### wiki package - content management
manages wikis: pages, categories, images, config

**core functionality:**
- **page operations** - create/edit/delete, git version control
- **category system** - hierarchical organization with auto-tracking
- **image processing** - server-side resizing, retina support, caching
- **model system** - reusable content templates/structured data
- **configuration** - layered config (system → wiki → page)

**locking coordination:**
- **cross-process locking** - coordinates cli/webserver/admin
- **hierarchical locks** - file → page → wiki level coordination
- **automatic cleanup** - handles stale locks/process failures

**key files:**
- `wiki.go` - main wiki type/operations
- `page.go` - page discovery/rendering/caching
- `category.go` - hierarchical category management
- `image.go` - advanced image processing pipeline
- `lock.go` - wiki-level locking coordination

### webserver package - http interface
serves wikis via http: viewing, editing, multi-site support

**features:**
- **multi-wiki hosting** - multiple wikis with virtual hosts
- **template system** - go templates with wiki content integration
- **static asset optimization** - efficient resource serving/caching
- **auth integration** - session-based auth/permission checking

**key files:**
- `webserver.go` - main http server setup/routing
- `wiki.go` - wiki-specific http operations/template rendering
- `template.go` - template system with wiki data integration

### adminifier package - web administration
comprehensive web admin interface for server and wiki management

**server-level admin:**
- **multi-site management** - overview/control of all wikis
- **user management** - authentication/permission administration
- **route inspector** - live view of registered http routes
- **help system** - integrated documentation

**wiki-level admin:**
- **dashboard** - real-time status with error/warning reporting
- **content management** - full crud: pages/images/models
- **category management** - nested organization with bulk operations
- **live editing** - ace editor with syntax highlighting/auto-save
- **settings** - live configuration editing with validation

**architecture:**
- **ajax-based ui** - dynamic content loading without page refresh
- **shared templates** - unified auth forms across systems
- **permission integration** - role-based access control

**key files:**
- `adminifier.go` - main admin interface setup
- `wiki.go` - wiki administration handlers/templates
- `handlers.go` - http handlers for admin operations

### authenticator package - security
user authentication/authorization with role-based access control

**features:**
- **secure authentication** - bcrypt password hashing (configurable cost)
- **session management** - secure session cookies with csrf protection
- **role-based permissions** - hierarchical roles with inheritance
- **user mapping** - server users map to different wiki usernames
- **cli tools** - comprehensive command-line user management

**key files:**
- `authenticator.go` - core authentication logic
- `user.go` - user management with secure password handling
- `role.go` - role-based permission system

### lock package - coordination
file-based cross-process locking for safe concurrent operations

**features:**
- **cross-process coordination** - works between cli/webserver/admin
- **timeout handling** - configurable timeouts prevent deadlocks
- **stale lock detection** - automatic cleanup of abandoned locks
- **hierarchical locking** - supports page-level and wiki-level operations

## supporting components

### pregenerate package - background processing
intelligent background page rendering with configurable performance

**features:**
- **dual-queue system** - priority vs background processing
- **rate limiting** - configurable 2-1000 pages/sec
- **dependency tracking** - understands page relationships for efficient updates
- **statistics** - comprehensive performance metrics/generation times

### monitor package - file watching
real-time file system monitoring with automatic page regeneration

**features:**
- **fsnotify integration** - efficient cross-platform file watching
- **automatic regeneration** - detects changes, triggers background processing
- **debouncing** - intelligent handling of rapid file changes
- **coordination** - integrates with locking system avoiding conflicts

### resources package - shared assets
embedded resource management eliminating duplication across components

**features:**
- **go:embed integration** - all assets bundled in binary
- **shared templates** - unified ui components across adminifier/webserver
- **consistent styling** - shared css/javascript libraries
- **embedded wikis** - base wiki templates/help documentation

### markdown package - format integration
sophisticated markdown to quiki translation for content integration

**features:**
- **configurable processing** - extensive flag system for customization
- **security controls** - optional html filtering for safety
- **format preservation** - maintains formatting intent during translation
- **metadata handling** - preserves document metadata during conversion

## data flow patterns

### request processing
```
http request → router → middleware → handler → wiki resolver → 
page parser → block processing → template rendering → response
```

### content generation
```
source file → file monitor → parser queue → tokenization → 
block recognition → variable resolution → html generation → cache
```

### background processing
```
file change → monitor detection → pregeneration queue → 
lock acquisition → async processing → cache invalidation → 
dependent page updates
```

## architectural strengths

### modular design
- **interface-based** - extensive go interfaces for extensibility
- **plugin architecture** - dynamic block type registration
- **separation of concerns** - clear boundaries: parsing/storage/serving

### performance optimization
- **intelligent caching** - multi-level caching with dependency tracking
- **lazy loading** - deferred loading of expensive resources
- **concurrent processing** - configurable parallelism for performance

### reliability
- **graceful degradation** - system continues despite non-critical errors
- **comprehensive locking** - prevents corruption during concurrent access
- **error recovery** - automatic recovery for transient failures

### security
- **input validation** - comprehensive sanitization throughout
- **access control** - role-based permissions with audit capabilities
- **secure sessions** - proper session management with csrf protection

## key file locations
```
wikifier/           # core parsing engine
wiki/              # content management  
webserver/         # http interface
adminifier/        # web administration
authenticator/     # security and auth
lock/              # coordination primitives
pregenerate/       # background processing
monitor/           # file watching
resources/         # embedded assets
```