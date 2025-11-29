# Enhanced Authorization System

## Tổng Quan

Hệ thống phân quyền nâng cao với cấu trúc phân cấp rõ ràng để quản lý quyền hạn chi tiết theo Module, Department, Service và Scope.

## Kiến Trúc Phân Quyền

### 1. **Module** (Mô-đun)
- Đại diện cho các khu vực chức năng lớn của hệ thống
- Ví dụ: `admin`, `crm`, `content`, `hr`, `finance`
- Mỗi module có thể chứa nhiều departments

**Thuộc tính:**
- `code`: Mã định danh duy nhất (e.g., "crm")
- `name`: Tên module
- `display_name`: Tên hiển thị
- `icon`, `color`: Giao diện
- `order`: Thứ tự hiển thị
- `is_active`: Trạng thái hoạt động
- `is_system`: Module hệ thống (không thể xóa)

### 2. **Department** (Phòng ban)
- Đại diện cho các phòng ban/bộ phận trong tổ chức
- Ví dụ: `sales`, `editorial`, `it`, `accounting`
- Thuộc về một Module cụ thể
- Hỗ trợ cấu trúc phân cấp (parent-child)

**Thuộc tính:**
- `module_id`: Thuộc module nào
- `code`: Mã định danh duy nhất
- `parent_id`: Phòng ban cha (cho cấu trúc phân cấp)
- `manager_id`: Người quản lý phòng ban

### 3. **Service** (Dịch vụ)
- Đại diện cho các chức năng/dịch vụ cụ thể
- Ví dụ: `user_management`, `customer_management`, `post_management`
- Thuộc về một Department cụ thể
- Mỗi service thường tương ứng với một API endpoint prefix

**Thuộc tính:**
- `department_id`: Thuộc department nào
- `code`: Mã định danh duy nhất
- `endpoint`: API endpoint prefix (e.g., "/api/v1/users")

### 4. **Scope** (Phạm vi)
- Định nghĩa phạm vi áp dụng của quyền
- 4 levels: `organization`, `department`, `team`, `personal`
- Priority càng cao = phạm vi càng rộng

**Levels:**
- **Organization** (100): Toàn tổ chức
- **Department** (50): Trong phòng ban
- **Team** (25): Trong nhóm
- **Personal** (10): Chỉ tài nguyên của bản thân

### 5. **EnhancedPermission** (Quyền nâng cao)
- Kết hợp tất cả các yếu tố trên
- Format: `module:department:service:scope:resource:action`
- Ví dụ: `crm:sales:customers:org:customers:read`

**Actions:**
- `create`, `read`, `update`, `delete`
- `execute`, `manage`, `approve`, `publish`
- `export`, `import`

## Ví Dụ Cụ Thể

### Ví dụ 1: Quản lý Khách hàng

```
Module: crm
  └─ Department: sales
      └─ Service: customers
          ├─ Permission: crm:sales:customers:org:customers:read
          │   → Xem tất cả khách hàng trong tổ chức
          ├─ Permission: crm:sales:customers:dept:customers:read
          │   → Xem khách hàng trong phòng sales
          └─ Permission: crm:sales:customers:personal:customers:read
              → Chỉ xem khách hàng được assign cho mình
```

### Ví dụ 2: Quản lý Bài viết

```
Module: content
  └─ Department: editorial
      └─ Service: posts
          ├─ Permission: content:editorial:posts:org:posts:publish
          │   → Publish bất kỳ bài viết nào
          ├─ Permission: content:editorial:posts:personal:posts:publish
          │   → Chỉ publish bài viết của mình
          ├─ Permission: content:editorial:posts:org:posts:update
          │   → Sửa bất kỳ bài viết nào
          └─ Permission: content:editorial:posts:personal:posts:update
              → Chỉ sửa bài viết của mình
```

### Ví dụ 3: Quản lý User

```
Module: admin
  └─ Department: system
      └─ Service: users
          ├─ Permission: admin:system:users:org:users:create
          │   → Tạo user mới
          ├─ Permission: admin:system:users:org:users:delete
          │   → Xóa bất kỳ user nào
          └─ Permission: admin:system:users:personal:users:update
              → Chỉ cập nhật profile của mình
```

## Cách Sử Dụng

### 1. Tạo Module mới

```go
module := &domain.Module{
    Code:        "hr",
    Name:        "Human Resources",
    DisplayName: "HR Management",
    Description: "Employee and HR management",
    Icon:        "users-cog",
    Color:       "#3498db",
    Order:       4,
    IsActive:    true,
    IsSystem:    false,
}
err := moduleUseCase.CreateModule(ctx, module)
```

### 2. Tạo Department

```go
department := &domain.Department{
    ModuleID:    hrModuleID,
    Code:        "recruitment",
    Name:        "Recruitment",
    DisplayName: "Recruitment Department",
    Description: "Hiring and recruitment",
    IsActive:    true,
}
err := departmentUseCase.CreateDepartment(ctx, department)
```

### 3. Tạo Service

```go
service := &domain.Service{
    DepartmentID: recruitmentDeptID,
    Code:         "candidates",
    Name:         "Candidate Management",
    DisplayName:  "Candidate Management Service",
    Description:  "Manage job candidates",
    Endpoint:     "/api/v1/candidates",
    IsActive:     true,
}
err := serviceUseCase.CreateService(ctx, service)
```

### 4. Tạo Permission

```go
permission := &domain.EnhancedPermission{
    ModuleID:     hrModuleID,
    DepartmentID: recruitmentDeptID,
    ServiceID:    candidatesServiceID,
    ScopeID:      orgScopeID,
    Resource:     "candidates",
    Action:       domain.ActionRead,
    DisplayName:  "View All Candidates",
    Description:  "View all candidates in organization",
    IsSystem:     false,
}
err := permissionUseCase.CreatePermission(ctx, permission)
```

### 5. Gán Permission cho Role

```go
permissionIDs := []uint{perm1ID, perm2ID, perm3ID}
err := permissionUseCase.AssignPermissionsToRole(ctx, roleID, permissionIDs)
```

### 6. Kiểm tra Permission

```go
// Kiểm tra bằng code
hasPermission, err := permissionUseCase.CheckUserPermission(ctx, userID, "crm:sales:customers:org:customers:read")

// Hoặc kiểm tra bằng components
hasPermission, err := permissionUseCase.CheckUserPermissionByComponents(
    ctx, userID, "crm", "sales", "customers", "org", "customers", "read"
)
```

## Migration & Seeding

### Database Tables

Hệ thống tạo các bảng sau:
- `modules`: Quản lý modules
- `departments`: Quản lý departments
- `services`: Quản lý services
- `scopes`: Quản lý scopes
- `enhanced_permissions`: Quản lý permissions
- `role_enhanced_permissions`: Mapping role-permission
- `user_enhanced_permissions`: Direct user permissions

### Seed Data

Khi chạy migration, hệ thống tự động seed:

**Modules:**
- `admin` - System Administration
- `crm` - Customer Relationship Management
- `content` - Content Management

**Departments:**
- `system` (admin)
- `sales` (crm)
- `editorial` (content)

**Services:**
- `users`, `roles`, `permissions` (system)
- `customers` (sales)
- `posts`, `media` (editorial)

**Scopes:**
- `org` - Organization (priority: 100)
- `dept` - Department (priority: 50)
- `team` - Team (priority: 25)
- `personal` - Personal (priority: 10)

**Enhanced Permissions:**
- 30+ permissions được tạo tự động
- Tất cả được gán cho role `super_admin`

## API Endpoints (Sẽ được implement)

### Modules
- `GET /api/v1/modules` - List modules
- `GET /api/v1/modules/:id` - Get module
- `POST /api/v1/modules` - Create module
- `PUT /api/v1/modules/:id` - Update module
- `DELETE /api/v1/modules/:id` - Delete module

### Departments
- `GET /api/v1/departments` - List departments
- `GET /api/v1/departments/:id` - Get department
- `GET /api/v1/modules/:moduleId/departments` - List by module
- `POST /api/v1/departments` - Create department
- `PUT /api/v1/departments/:id` - Update department
- `DELETE /api/v1/departments/:id` - Delete department

### Services
- `GET /api/v1/services` - List services
- `GET /api/v1/services/:id` - Get service
- `GET /api/v1/departments/:deptId/services` - List by department
- `POST /api/v1/services` - Create service
- `PUT /api/v1/services/:id` - Update service
- `DELETE /api/v1/services/:id` - Delete service

### Scopes
- `GET /api/v1/scopes` - List scopes
- `GET /api/v1/scopes/:id` - Get scope

### Enhanced Permissions
- `GET /api/v1/enhanced-permissions` - List permissions
- `GET /api/v1/enhanced-permissions/:id` - Get permission
- `POST /api/v1/enhanced-permissions` - Create permission
- `PUT /api/v1/enhanced-permissions/:id` - Update permission
- `DELETE /api/v1/enhanced-permissions/:id` - Delete permission
- `POST /api/v1/roles/:roleId/enhanced-permissions` - Assign to role
- `POST /api/v1/users/:userId/enhanced-permissions` - Grant to user
- `GET /api/v1/users/:userId/enhanced-permissions` - Get user permissions
- `POST /api/v1/users/:userId/check-permission` - Check permission

## Lợi Ích

### 1. **Phân quyền Chi Tiết**
- Kiểm soát quyền ở nhiều cấp độ: organization, department, team, personal
- Dễ dàng implement row-level security

### 2. **Dễ Mở Rộng**
- Thêm module mới không ảnh hưởng code cũ
- Thêm department, service mới rất đơn giản
- Tạo permission mới chỉ cần vài dòng code

### 3. **Rõ Ràng & Dễ Hiểu**
- Cấu trúc phân cấp rõ ràng
- Permission code tự mô tả: `crm:sales:customers:org:customers:read`
- Dễ debug và audit

### 4. **Tương Thích Với Hệ Thống Lớn**
- Phù hợp cho multi-tenant
- Hỗ trợ organizational hierarchy
- Scale tốt khi hệ thống phát triển

### 5. **Audit & Compliance**
- Track được ai có quyền gì
- Biết permission được grant bởi ai, khi nào
- Hỗ trợ expiration cho temporary permissions

## Next Steps

1. ✅ Domain models created
2. ✅ Repository interfaces created
3. ✅ Repository implementations created
4. ✅ Migration & seed data created
5. ⏳ Use case implementations (next)
6. ⏳ HTTP handlers (next)
7. ⏳ Router integration (next)
8. ⏳ Middleware for permission checking (next)
9. ⏳ Frontend integration (next)

## Notes

- Legacy `Permission` model vẫn được giữ lại để backward compatibility
- Có thể migrate dần từ legacy sang enhanced permission
- `super_admin` role tự động có tất cả permissions
- System entities (is_system=true) không thể xóa
