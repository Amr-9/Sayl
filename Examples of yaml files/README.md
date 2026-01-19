# Sayl YAML Configuration Examples

This directory contains example configuration files for the Sayl load testing tool. Each file demonstrates different features and scenarios.

## File List

### Basic Scenarios
- **01_basic_get.yaml**: A simple GET request configuration.
- **02_post_json.yaml**: A POST request with a JSON payload, using the `body_json` field.
- **03_post_raw_body.yaml**: A POST request using a raw string for the body, useful for non-JSON content.
- **09_post_from_file.yaml**: sending a request body from an external file (e.g., large JSON or binary).

### Advanced Load Patterns
- **04_load_stages.yaml**: How to configure load ramping (stages) to simulate traffic spikes.
- **05_data_loader.yaml**: Using CSV files to feed dynamic data into your requests (e.g., using different user IDs).
- **12_multiple_csv_sources.yaml**: Using multiple CSV files simultaneously (e.g., Users and Products).

### Scenario Chaining
- **06_scenario_chain.yaml**: A multi-step scenario (e.g., Login -> Get Profile) using chained requests and variable extraction.
- **13_extract_headers.yaml**: Extracting values from response headers (e.g., `Set-Cookie`, `ETag`) for use in subsequent steps.
- **20_mixed_crud_scenario.yaml**: A full lifecycle scenario: Create -> Read -> Update -> Delete.

### Advanced Configuration & Auth
- **07_auth_headers.yaml**: Adding Authorization headers and other custom headers.
- **08_advanced_config.yaml**: A comprehensive configuration using advanced settings like timeouts, keep-alive, and connection pooling.

### Specific Request Types
- **10_graphql_query.yaml**: Sending GraphQL queries and variables.
- **11_form_urlencoded.yaml**: Sending `application/x-www-form-urlencoded` form data.
- **14_put_update.yaml**: Using the PUT method for resource updates.
- **15_delete_resource.yaml**: Using the DELETE method for removing resources.
- **16_patch_partial_update.yaml**: Using the PATCH method for partial updates.
- **17_complex_json_body.yaml**: Constructing deeply nested and complex JSON bodies.
- **18_query_params.yaml**: Handling dynamic query parameters in URLs.

### Variable System
- **19_variables_demo.yaml**: Showcase of all built-in dynamic variables (`uuid`, `timestamp`, `random_int`).

## Key Concepts

### Target
The `target` section defines the endpoint you are testing.
- `url`: The full URL.
- `method`: GET, POST, PUT, DELETE, etc.
- `body_json`: A structured object that Sayl converts to JSON automatically.

### Load
The `load` section controls how hard you hit the target.
- `duration`: How long the test runs (e.g., "30s", "5m").
- `rate`: Requests per second (RPS).
- `concurrency`: Number of parallel workers.
- `stages`: Define changes in load over time (e.g., ramp up from 10 to 100 RPS).

### Variables & Templating
You can use variables in your requests using `{{ variable_name }}` syntax.
- **Dynamic Variables**:
    - `{{ uuid }}`: Generates a random UUID.
    - `{{ random_int }}`: Generates a random integer.
    - `{{ timestamp }}`: Current Unix timestamp.
- **Data Source Variables**:
    - If you load a CSV named `users` with a column `email`, accessing it is `{{ users.email }}`.
- **Extracted Variables**:
    - In chained scenarios, you can extract values from a response and use them in the next step.
