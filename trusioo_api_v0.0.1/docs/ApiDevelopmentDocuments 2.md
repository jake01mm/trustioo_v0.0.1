

## 1.1. 测卡接口

使用api发送请求，检测卡片使用状态

## 1.2. 请求方法

### 1.2.1. Http Method

  

```
 POST(application/json)
```

### 1.2.2. Http 返回格式

​    

```
JSON
```

### 1.2.3. URL

 

```url
 http://{$host_name}/api/userApiManage/checkCard
```

注: host_name 为ck象服务服务器地址

### 1.2.4. Http 请求参数说明

注：appId与appSecret 在平台申请获取

#### 请求头：

| 参数  | 类型   | 是否必须 | 说明         | 示例                |
| ----- | ------ | -------- | ------------ | ------------------- |
| appId | String | 必须     | 账户申请获取 | 2410101358207987266 |

#### 请求参数（json）：

| 参数 | 类型   | 是否必须 | 说明          |
| :--- | :----- | :------- | :------------ |
| data | String | 必须     | AES加密字符串 |

### 1.2.5. data参数说明

字符串data由json字符串通过AES加密生成

```json
{
  "cards": [
    "X123123123123123" #卡号  如有PIN码 使用-隔开 如：X123123123123123-12312312（丝芙兰、nike、nd）
  ],
  "timestamp": "1728629545", # 请求发送时间戳
  "sign": "",		#签名
  "productMark": "iTunes",   #产品类型（下方有字典匹配）
  "regionId": 1,		# 国家或地区编号 例：1（下方有字典匹配）
  "regionName": "美国",		# 国家或地区编号 例：美国（下方有字典匹配）
  "autoType": 0				#是否自动测卡（仅苹果测卡使用）0指定国家 1自动识别
}

```



### 1.2.6. Http 返回结果说明

| 字段 | 类型    | 描述       | 示例      |      |
| :--- | :------ | :--------- | :-------- | ---- |
| code | Int     | 返回状态码 | 如: 200   |      |
| msg  | String  | 提示信息   | 默认为 "" |      |
| data | Boolean | 请求结果   | 布尔类型  |      |

```json
示例
成功：
{
    "code": 200,
    "msg": "",
    "data": true
}
失败：
{
    "code": 500,
    "msg": "验签失败",   #错误原因视msg而定
    "data": ""
}
```



## 1.3. 请求示例

### 1.3.1. Java调用示例

```java
		//  苹果测卡示例  
		// List<String> cards = bo.getCards();
		JSONArray cards = JSONUtil.createArray();//使用JSONArray
        cards.add("X123123123123122");  	//卡号信息
        Map<String, Object> map = new HashMap<>();
        map.put("cards", cards);
        map.put("productMark", "iTunes");
        map.put("regionId", 1);
        map.put("regionName", "美国");
        map.put("autoType", 0);
        map.put("timestamp", DateUtil.currentSeconds());
        String sign = SignUtil.sign(map, appSecret);
        map.put("sign", sign);
        String encryptData = DesUtil.getEncryptData(appSecret, JSONUtil.toJsonStr(map));
        JSONObject jsonObject = new JSONObject();
        jsonObject.set("data", encryptData);
        HttpResponse execute = HttpRequest.post(url)
                .header("Content-Type", "application/json")
                .header("appId", appId)
                .contentType("application/json")
                .body(JSONUtil.toJsonStr(jsonObject))
                .execute();
        String body = execute.body();
```

### 1.3.2. 成功返回示例

```json
{
    "code": 200,
    "msg": "",
    "data": true   # false为失败
}
```
### 1.3.3. 失败返回示例（及说明）

  

```json
{
    "code": 500,
    "msg": "验签失败",   #错误原因视msg而定
    "data": ""
}
```



------



## 2.1. 接口查询测卡结果

使用api发送请求，检测卡片测卡结果。

## 2.2. 请求方法

### 2.2.1. Http Method

  

```
 POST(application/json)
```

### 2.2.2. Http 返回格式

​    

```
JSON
```

### 2.2.3. URL

 

```
 http://{$host_name}/api/userApiManage/checkCardResult
```

注: host_name 为ck象服务服务器地址

### 2.2.4. Http 请求参数说明

#### 请求头：

| 参数  | 类型   | 是否必须 | 说明         | 示例                |
| ----- | ------ | -------- | ------------ | ------------------- |
| appId | String | 必须     | 账户申请获取 | 2410101358207987266 |

#### 请求参数（json）：

| 参数 | 类型   | 是否必须 | 说明          |
| :--- | :----- | :------- | :------------ |
| data | String | 必须     | AES加密字符串 |

### 2.2.5. data参数说明

字符串data由json字符串通过AES加密生成

```json
 {
  "timestamp": "1728629545", # 请求发送时间戳
  "sign": "",		#签名
  "productMark": "iTunes",   #产品类型（下方有字典匹配） 必填
  "cardNo": "X123123123123123", #卡号  必填
  "pinCode": "" #PIN码  含有PIN码的必填（丝芙兰、nike、nd）
}
```



### 2.2.6. Http 返回结果说明

| 字段 | 类型   | 描述       | 示例                     |      |
| :--- | :----- | :--------- | :----------------------- | ---- |
| code | Int    | 返回状态码 | 如: 200                  |      |
| msg  | String | 提示信息   | 默认为 ""                |      |
| data | String | 请求结果   | code:200 为AES加密的结果 |      |

```json
示例
成功：
{
    "code": 200,
    "msg": "",
    "data": "dsadsad"  #AES加密结果   返回结果后 使用AES解密
}
失败：
{
    "code": 500,
    "msg": "验签失败",   #错误原因视msg而定
    "data": ""
}

```



## 2.3. 请求示例

### 2.3.1. Java调用示例

```java
	//  苹果测卡示例  
	String cardNo = "X123123123123122";
    String pinCode = "";
    Map<String, Object> map = new HashMap<>();
    map.put("cardNo", cardNo);
    map.put("pinCode", pinCode);
    map.put("productMark", "iTunes");
    map.put("timestamp", DateUtil.currentSeconds());
    String sign = SignUtil.sign(map, appSecret);
    map.put("sign", sign);
    String encryptData = DesUtil.getEncryptData(appSecret, JSONUtil.toJsonStr(map));
    JSONObject jsonObject = new JSONObject();
    jsonObject.set("data", encryptData);
//    String params = HttpUtil.toParams(jsonObject);
    HttpResponse execute = HttpRequest.post(url)
            .header("Content-Type", "application/json")
            .header("appId", appId)
            .contentType("application/json")
            .body(JSONUtil.toJsonStr(jsonObject))
            .execute();
    String body = execute.body();
    JSONObject parseObj = JSONUtil.parseObj(body);
    Integer code = parseObj.getInt("code");
    if(code == 200){
        String data = parseObj.getStr("data");
        String decryptData = DesUtil.getDecryptData(appSecret, data);
        parseObj.set("data", decryptData);
        return JSONUtil.toJsonStr(parseObj);
    }else {
        return parseObj.getStr("msg");
    }
```

### 2.3.2. 成功返回示例

```json
{
    "code": 200,
    "msg": "",
    "data": "dsadsad"  #AES加密结果   返回结果后 使用AES解密
}

# 解密结果：
{
  "cardNo": "",  #请求的卡号
  "status": 0, #状态 '0-等待检测，1-测卡中，2-有效，3-无效，4-已兑换，5-检测失败，6-点数不足
  "pinCode": "",#PIN码
  "message": "",#检测结果信息（错误信息）
  "checkTime": "",#检测时间
  "regionName": "",# 卡种国家（部分含有）
  "regionId": 0 #卡种国家编号（部分含有）
}
```

### 2.3.3. 失败返回示例（及说明）

  

```json
{
    "code": 500,
    "msg": "验签失败",   #错误原因视msg而定
    "data": ""
}
```

------



## 3 签名+加密公共方法

### 3.1 签名（SignUtil）

- 【对所有API请求参数（包括公共参数和业务参数，但除去sign参数和byte[]类型的参数），根据参数名称的ASCII码表的顺序排序。如：foo=1, bar=2, foo_bar=3, foobar=4排序后的顺序是bar=2, foo=1, foo_bar=3, foobar=4。
- 将排序好的参数名和参数值拼装在一起，根据上面的示例得到的结果为：bar2foo1foo_bar3foobar4。
- 把拼装好的字符串采用utf-8编码，在拼装的字符串前后加上app的secret后，使用MD5算法进行摘要，如：md5(secret+bar2foo1foo_bar3foobar4+secret)；



```java
import lombok.extern.slf4j.Slf4j;
import org.elvis.common.exception.ApiException;
import java.io.UnsupportedEncodingException;
import java.security.MessageDigest;
import java.security.NoSuchAlgorithmException;
import java.util.Arrays;
import java.util.Map;  
// 签名
    public static String sign(Map<String, Object> params, String appSecret) {
        try {
            String[] keys = params.keySet().toArray(new String[0]);
            Arrays.sort(keys);
            StringBuilder query = new StringBuilder();
            query.append(appSecret);
            for (String key : keys) {
                String value = params.get(key).toString();
                query.append(key).append(value);
            }
            query.append(appSecret);
            byte[] md5byte = encryptMD5(query.toString());
            String signStr = byte2hex(md5byte);
            return signStr;
        }
        catch (Exception e) {
            log.error("签名生成失败", e);
            return null;
        }
    }
// 验签
  public static boolean verifySign(Map<String, Object> params, String appSecret, String signToVerify) {
        try {
            params.remove("sign");
            String[] keys = params.keySet().toArray(new String[0]);
            Arrays.sort(keys);
            StringBuilder query = new StringBuilder();
            query.append(appSecret);
            for (String key : keys) {
                String value = params.get(key).toString();
                query.append(key).append(value);
            }
            query.append(appSecret);
            byte[] md5byte = encryptMD5(query.toString());
            String sign = byte2hex(md5byte);
            return sign.equals(signToVerify);
        }
        catch (Exception e) {
            throw new ApiException("验签失败");
        }
    }

    // byte数组转成16进制字符串
    public static String byte2hex(byte[] bytes) {
        StringBuilder sign = new StringBuilder();
        for (int i = 0; i < bytes.length; i++) {
            String hex = Integer.toHexString(bytes[i] & 0xFF);
            if (hex.length() == 1) {
                sign.append("0");
            }
            sign.append(hex.toLowerCase());
        }
        return sign.toString();
    }

    // Md5摘要
    public static byte[] encryptMD5(String data) throws NoSuchAlgorithmException, UnsupportedEncodingException {
        MessageDigest md5 = MessageDigest.getInstance("MD5");
        return md5.digest(data.getBytes("UTF-8"));
    }
```

### 3.2 AES加解密

```java

    /**
     * 获取加密后信息
     * @param plainText 明文
     * @return 加密后信息
     */
    public static String getEncryptData(String appSecret, String plainText) {
        DES des = SecureUtil
                .des(SecureUtil.generateKey(SymmetricAlgorithm.DES.getValue(), appSecret.getBytes()).getEncoded());
        return des.encryptHex(plainText); // 加密为16进制
    }

    /**
     * 获取解密后信息
     * @param cipherText 密文
     * @return 解密后信息
     */
    public static String getDecryptData(String appSecret, String cipherText) {
        if (StrUtil.isBlank(appSecret)) {
            throw new ApiException("解密失败，解密key为空");
        }
        if (StrUtil.isBlank(cipherText)) {
            throw new ApiException("解密失败，解密内容为空");
        }
        DES des = SecureUtil
                .des(SecureUtil.generateKey(SymmetricAlgorithm.DES.getValue(), 		appSecret.getBytes()).getEncoded());
        return des.decryptStr(cipherText);
    }
```

------



## 4. 各测卡功能参数解析

### 4.1 通用必填参数（**productMark**）

#### productMark：必填参数

| 功能       | 值      | 说明 |
| ---------- | ------- | ---- |
| 丝芙兰测卡 | sephora |      |
| 雷蛇测卡   | Razer   |      |
| 苹果测卡   | iTunes  |      |
| 亚马逊测卡 | amazon  |      |
| XBOX测卡   | xBox    |      |
| NIKE测卡   | nike    |      |
| ND测卡     | nd      |      |

### 4.2 丝芙兰测卡

#### 4.2.1 请求参数

```json
#请求body：
{
    "cards": [
    "1123123123123123-12312312" #卡号  如有PIN码 使用-隔开 如：X123123123123123-12312312
  ],
    "productMark":"sephora"
}
#卡号：16位数字
#PIN码：8位数字
```

### 4.3 雷蛇测卡

#### 4.3.1 请求参数

```json
#请求body：
{
    "cards": [
    "1123123123123123" #卡号  
  ],
    "productMark":"Razer",
    "regionId":"" #测卡国家或区域编码
}
#国家对应regionId
 [{"regionId":12,"chName":"美国"},{"regionId":6,"chName":"澳大利亚"},{"regionId":13,"chName":"巴西"},{"regionId":26,"chName":"柬埔寨"},{"regionId":20,"chName":"加拿大"},{"regionId":25,"chName":"智利"},{"regionId":22,"chName":"哥伦比亚"},{"regionId":17,"chName":"香港特别行政区"},{"regionId":4,"chName":"印度"},{"regionId":7,"chName":"印度尼西亚"},{"regionId":27,"chName":"日本"},{"regionId":1,"chName":"马来西亚"},{"regionId":19,"chName":"缅甸"},{"regionId":15,"chName":"新西兰"},{"regionId":29,"chName":"巴基斯坦"},{"regionId":8,"chName":"菲律宾"},{"regionId":5,"chName":"新加坡"},{"regionId":18,"chName":"土耳其"},{"regionId":33,"chName":"越南"},{"regionId":2,"chName":"其他"},{"regionId":28,"chName":"其他（中文）"},{"regionId":21,"chName":"墨西哥"}]
```

### 4.4 苹果测卡

#### 4.4.1 请求参数

```json
#请求body：
{
    "cards": [
    "1123123123123123" #卡号  
  ],
    "productMark":"iTunes",
    "regionId":"2", #测卡国家或区域编码
    "regionName":"美国",
    "autoType":0    #是否自动测卡（仅苹果测卡使用）0指定国家 1自动识别
}
#国家对应regionId
[{"id":1,"regionName":"英国"},{"id":2,"regionName":"美国"},{"id":3,"regionName":"德国"},{"id":4,"regionName":"澳大利亚"},{"id":5,"regionName":"加拿大"},{"id":6,"regionName":"日本"},{"id":8,"regionName":"西班牙"},{"id":9,"regionName":"意大利"},{"id":10,"regionName":"法国"},{"id":11,"regionName":"爱尔兰"},{"id":12,"regionName":"墨西哥"}]
```

### 4.5 亚马逊测卡  

#### 4.5.1 请求参数

```json
#请求body：
{
    "cards": [
    "1123123123123123" #卡号  卡片格式:电子码是14位卡号，实体卡是15位卡号
  ],
    "productMark":"amazon",
    "regionId":"2" #测卡国家或区域编码
}
#国家对应regionId
[{"regionId":2,"regionName":"美亚/加亚"},{"regionId":1,"regionName":"欧盟区"}]
# 欧盟区：支持英国、德国、荷兰、西班牙、法国、奥 地利、丹麦、芬兰、希腊、意大利、波兰、 葡萄牙、瑞典
```

### 4.6 XBOX测卡 

#### 4.6.1 请求参数

```json
#请求body：
{
    "cards": [
    "1231231231231231231231231" #卡号  卡片格式:25个字符
  ],
    "productMark":"xBox",
    "regionName":"美国" #测卡国家或区域
}
#国家
["美国","加拿大","英国","澳大利亚","新西兰","新加坡","韩国","墨西哥","瑞典","哥伦比亚","阿根廷","尼日利亚","香港特别行政区","挪威","波兰","德国"]
```

### 4.7 NIKE测卡 

#### 4.7.1 请求参数

```json
#请求body：
{
    "cards": [
    "1231231231231231231-231231" #卡号  卡片格式：卡号固定19位，pin码固定6位 {codeNo}-{pinCode}
  ],
    "productMark":"nike"
}
```

### 4.8 ND测卡 

#### 4.8.1 请求参数

```json
#请求body：
{
    "cards": [
    "1231231231231231-23123123" #卡号  卡片格式：卡号固定16位，pin码固定8位，可以连续输 {codeNo}-{pinCode}
  ],
    "productMark":"nd"
}
```

