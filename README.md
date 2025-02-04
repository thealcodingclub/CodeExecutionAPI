# Code Execution API

This API is used to execute code snippets in different languages. Given below are the details of this API

## Route: /execute

This route takes 4 fields:

1. `langauge`: The language of the code snippet.
    <details>
    <summary>Click to see supported languages</summary>

    - python
    - rust
    - cpp
    - c

    </details>

2. `code`: The code snippet to be executed.
3. `timeout`: The maximum time in seconds for which the code should run. If the code runs for more than this time, it will be terminated. 
    - default: **5 seconds**
    - max: **60 seconds**
4. `max_memory`: The maximum memory in KB (kilobytes) that the code can use. If the code uses more memory than this, it will be terminated.
    - default: **32768KB** (or 32MB)
    - max: **131072KB** (or 128MB)

---

#### Request body format (Example 1):

```json
{
    "language": "python",
    "code": "print('Hello World')"
}
```
<details>
<summary>Click to copy curl command</summary>

```bash
curl --location 'localhost:8080/execute' \
--header 'Content-Type: application/json' \
--data '{
    "language": "python",
    "code": "print('\''Hello World'\'')"
}'
```

</details>

#### Response body format (Example 1):

```json
{
    "output": "Hello World\n",  // Output of the code
    "error": "",                // If any error occurs during execution
    "memory_used": "13808 KB",  // RAM used (in KB)
    "cpu_time": "125.034027ms"  // in Seconds
}
```

---

#### Request body format (Example 2):

```json
{
    "language": "python",
    "code": "import time\nprint('Hello World')\ntime.sleep(5)",
    "timeout": 2,       // in seconds (defaults to 5, max 60)
}
```

<details>
<summary>Click to copy curl command</summary>

```bash
curl --location 'localhost:8080/execute' \
--header 'Content-Type: application/json' \
--data '{
    "language": "python",
    "code": "import time\nprint('\''Hello World'\'')\ntime.sleep(5)",
    "timeout": 2       
}'
```

</details>

#### Response body format (Example 2):

```json
{
    "output": "",                   // No output is returned on timeout
    "error": "Execution Timed Out", // Error message in case of timeout
    "memory_used": "15856 KB",      // RAM used (in KB)
    "cpu_time": "2.000637132s"      // Time before code was terminated
}
```

---

#### Request body format (Example 3):

```json
{
    "language": "python",
    "code": "import random;[random.random() for x in range(10**7)]",
    "max_memory": 300000        // in KB (defaults to 32768, max 131072)
}
```

<details>
<summary>Click to copy curl command</summary>

```bash
curl --location 'localhost:8080/execute' \
--header 'Content-Type: application/json' \
--data '{
    "language": "python",
    "code": "import random;[random.random() for x in range(10**7)]",
    "max_memory": 300000
}'
```

</details>

#### Response body format (Example 3):

```json
{
    "output": "",
    "error": "Traceback (most recent call last):\n  File \"<string>\", line 1, in <module>\n  File \"<string>\", line 1, in <listcomp>\nMemoryError\n",
    "memory_used": "283964 KB",
    "cpu_time": "803.820445ms"
}
```

---

#### Request body format (Example 4):

```json
{
    "language": "python",
    "code": "a = input()\nprint(f'first value entered is {a}.')\nb=input()\nprint(f'second value entered is {b}.')",
    "inputs": [
        "bob",
        "alice"
    ]
}
```

<details>
<summary>Click to copy curl command</summary>

```bash
curl --location 'localhost:8080/execute' \
--header 'Content-Type: application/json' \
--data '{
    "language": "python",
    "code": "a = input()\nprint(f'\''first value entered is {a}.'\'')\nb=input()\nprint(f'\''second value entered is {b}.'\'')",
    "inputs": [
        "bob",
        "alice"
    ]
}'
```

</details>

#### Response body format (Example 4):

```json
{
    "output": "first value entered is bob.\nsecond value entered is alice.\n",
    "error": "",
    "memory_used": "13400 KB",
    "cpu_time": "135.266682ms"
}
```