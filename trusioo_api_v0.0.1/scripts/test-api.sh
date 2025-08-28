#!/bin/bash

# API测试脚本
# 用于测试Trusioo API的各个端点

set -e

BASE_URL="http://localhost:8080"
API_V1="$BASE_URL/api/v1"
AUTH_API="$API_V1/auth"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 输出函数
print_header() {
    echo -e "${BLUE}===========================================${NC}"
    echo -e "${BLUE} $1 ${NC}"
    echo -e "${BLUE}===========================================${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}ℹ️  $1${NC}"
}

# 测试函数
test_endpoint() {
    local method=$1
    local url=$2
    local data=$3
    local expected_status=$4
    local description=$5

    echo -e "\n${YELLOW}Testing:${NC} $description"
    echo -e "${YELLOW}URL:${NC} $method $url"
    
    if [ -n "$data" ]; then
        echo -e "${YELLOW}Data:${NC} $data"
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
                       -H "Content-Type: application/json" \
                       -d "$data" \
                       "$url")
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" "$url")
    fi

    # 分离响应体和状态码
    body=$(echo "$response" | head -n -1)
    status=$(echo "$response" | tail -n 1)

    echo -e "${YELLOW}Response Status:${NC} $status"
    echo -e "${YELLOW}Response Body:${NC}"
    echo "$body" | jq . 2>/dev/null || echo "$body"

    if [ "$status" = "$expected_status" ]; then
        print_success "Test passed (Status: $status)"
    else
        print_error "Test failed (Expected: $expected_status, Got: $status)"
    fi
}

# 主测试流程
main() {
    print_header "Trusioo API Test Suite"

    # 检查服务是否运行
    print_info "Checking if server is running..."
    if ! curl -s "$BASE_URL/ping" > /dev/null; then
        print_error "Server is not running at $BASE_URL"
        print_info "Please start the server first using: make dev"
        exit 1
    fi
    print_success "Server is running"

    # 1. 基础端点测试
    print_header "Basic Endpoints"
    
    test_endpoint "GET" "$BASE_URL/ping" "" "200" "Ping endpoint"
    test_endpoint "GET" "$BASE_URL/version" "" "200" "Version endpoint"
    test_endpoint "GET" "$API_V1" "" "200" "API v1 info"

    # 2. 健康检查测试
    print_header "Health Check Endpoints"
    
    test_endpoint "GET" "$BASE_URL/health" "" "200" "Overall health check"
    test_endpoint "GET" "$BASE_URL/health/database" "" "200" "Database health check"
    test_endpoint "GET" "$BASE_URL/health/redis" "" "200" "Redis health check"
    test_endpoint "GET" "$BASE_URL/health/api/v1" "" "200" "API v1 health check"
    test_endpoint "GET" "$BASE_URL/health/liveness" "" "200" "Liveness check"
    test_endpoint "GET" "$BASE_URL/health/readiness" "" "200" "Readiness check"

    # 3. 认证端点测试
    print_header "Authentication Endpoints"

    # 管理员登录测试
    admin_login_data='{
        "email": "admin@trusioo.com",
        "password": "admin123"
    }'
    test_endpoint "POST" "$AUTH_API/admin/login" "$admin_login_data" "200" "Admin login"

    # 用户注册测试
    user_register_data='{
        "email": "test@example.com",
        "password": "password123",
        "name": "Test User"
    }'
    test_endpoint "POST" "$AUTH_API/user/register" "$user_register_data" "201" "User registration"

    # 用户登录测试
    user_login_data='{
        "email": "test@example.com",
        "password": "password123"
    }'
    test_endpoint "POST" "$AUTH_API/user/login" "$user_login_data" "200" "User login"

    # 买家注册测试
    buyer_register_data='{
        "email": "buyer@company.com",
        "password": "password123",
        "company_name": "Test Company",
        "contact_name": "Test Contact",
        "phone": "+1234567890"
    }'
    test_endpoint "POST" "$AUTH_API/buyer/register" "$buyer_register_data" "201" "Buyer registration"

    # 买家登录测试
    buyer_login_data='{
        "email": "buyer@company.com",
        "password": "password123"
    }'
    test_endpoint "POST" "$AUTH_API/buyer/login" "$buyer_login_data" "200" "Buyer login"

    # 4. 错误处理测试
    print_header "Error Handling"
    
    test_endpoint "GET" "$BASE_URL/nonexistent" "" "404" "404 Not Found"
    test_endpoint "POST" "$BASE_URL/ping" "" "405" "405 Method Not Allowed"
    
    # 无效认证数据测试
    invalid_login_data='{
        "email": "invalid@example.com",
        "password": "wrongpassword"
    }'
    test_endpoint "POST" "$AUTH_API/admin/login" "$invalid_login_data" "401" "Invalid admin login"

    print_header "Test Summary"
    print_success "All tests completed!"
    print_info "Check the results above for any failed tests."
    print_info "You can also view the server logs with: make logs"
}

# 运行测试
main "$@"