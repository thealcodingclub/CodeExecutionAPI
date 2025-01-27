# Code Execution API

This API is used to execute code snippets in different languages. Given below are the details of this API

## Routes

### 1. /execute
Request Type: POST

#### Request body format (Example 1):

```json
{
    "language": "python",
    "code": "print('Hello World')"
}
```

#### Response body format (Example 1):

```json
{
    "output": "Hello World\n", // Output of the code
    "error": "",            // If any error occurs during execution
    "memory_used": "0.0",   // in Bytes
    "cpu_time": "0.0"       // in Seconds
}
```

#### Request body format (Example 2):

```json
{
    "language": "python",
    "code": "import time\nprint('Hello World')\ntime.sleep(5)",
    "timeout": 2,       // in seconds (defaults to 5)
}
```

#### Response body format (Example 2):

```json
{
    "output": "",           // Output of the code
    "error": "Execution Timed Out", // If any error occurs during execution
    "memory_used": "0.0",   // in Bytes
    "cpu_time": "2.0"       // in Seconds
}
```

#### Request body format (Example 3):

```json
{
    "language": "python",
    "code": "variable = [x for x in range(10**6)]",
    "memory_limit": 1000000, // in Bytes (defaults to 1000000)
}
```

#### Response body format (Example 3):

```json
{
    "output": "",           // Output of the code
    "error": "Memory Limit Exceeded", // If any error occurs during execution
    "memory_used": "1000000",   // in Bytes
    "cpu_time": "0.0"       // in Seconds
}
```