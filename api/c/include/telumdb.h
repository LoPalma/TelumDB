/*
 * TelumDB C Library
 * 
 * Copyright 2024 TelumDB Contributors
 * Licensed under the Apache License, Version 2.0
 */

#ifndef TELUMDB_H
#define TELUMDB_H

#include <stdint.h>
#include <stdbool.h>
#include <stddef.h>

#ifdef __cplusplus
extern "C" {
#endif

/* Version information */
#define TELUMDB_VERSION_MAJOR 0
#define TELUMDB_VERSION_MINOR 1
#define TELUMDB_VERSION_PATCH 0
#define TELUMDB_VERSION_STRING "0.1.0"

/* Error codes */
typedef enum {
    TELUMDB_OK = 0,
    TELUMDB_ERROR = -1,
    TELUMDB_ERROR_OUT_OF_MEMORY = -2,
    TELUMDB_ERROR_INVALID_PARAMETER = -3,
    TELUMDB_ERROR_CONNECTION_FAILED = -4,
    TELUMDB_ERROR_QUERY_FAILED = -5,
    TELUMDB_ERROR_TIMEOUT = -6,
    TELUMDB_ERROR_PROTOCOL_ERROR = -7,
    TELUMDB_ERROR_AUTH_FAILED = -8,
    TELUMDB_ERROR_NOT_CONNECTED = -9,
    TELUMDB_ERROR_ALREADY_CONNECTED = -10,
    TELUMDB_ERROR_TENSOR_SHAPE_MISMATCH = -11,
    TELUMDB_ERROR_TENSOR_TYPE_MISMATCH = -12,
    TELUMDB_ERROR_TENSOR_OUT_OF_BOUNDS = -13,
} telumdb_error_t;

/* Data types */
typedef enum {
    TELUMDB_TYPE_INT32 = 0,
    TELUMDB_TYPE_INT64 = 1,
    TELUMDB_TYPE_FLOAT32 = 2,
    TELUMDB_TYPE_FLOAT64 = 3,
    TELUMDB_TYPE_STRING = 4,
    TELUMDB_TYPE_BOOL = 5,
    TELUMDB_TYPE_BYTES = 6,
} telumdb_data_type_t;

/* Tensor data types */
typedef enum {
    TELUMDB_TENSOR_DTYPE_INT32 = 0,
    TELUMDB_TENSOR_DTYPE_INT64 = 1,
    TELUMDB_TENSOR_DTYPE_FLOAT32 = 2,
    TELUMDB_TENSOR_DTYPE_FLOAT64 = 3,
} telumdb_tensor_dtype_t;

/* Forward declarations */
typedef struct telumdb_connection telumdb_connection_t;
typedef struct telumdb_result telumdb_result_t;
typedef struct telumdb_tensor telumdb_tensor_t;
typedef struct telumdb_config telumdb_config_t;

/* Configuration structure */
struct telumdb_config {
    const char* host;
    int port;
    const char* database;
    const char* username;
    const char* password;
    int timeout_seconds;
    bool use_ssl;
    int max_connections;
};

/* Connection management */
telumdb_error_t telumdb_config_init(telumdb_config_t** config);
telumdb_error_t telumdb_config_set_host(telumdb_config_t* config, const char* host);
telumdb_error_t telumdb_config_set_port(telumdb_config_t* config, int port);
telumdb_error_t telumdb_config_set_database(telumdb_config_t* config, const char* database);
telumdb_error_t telumdb_config_set_credentials(telumdb_config_t* config, const char* username, const char* password);
telumdb_error_t telumdb_config_set_timeout(telumdb_config_t* config, int timeout_seconds);
telumdb_error_t telumdb_config_set_ssl(telumdb_config_t* config, bool use_ssl);
telumdb_error_t telumdb_config_free(telumdb_config_t* config);

telumdb_error_t telumdb_connect(const telumdb_config_t* config, telumdb_connection_t** connection);
telumdb_error_t telumdb_disconnect(telumdb_connection_t* connection);
bool telumdb_is_connected(const telumdb_connection_t* connection);
telumdb_error_t telumdb_ping(telumdb_connection_t* connection);
telumdb_error_t telumdb_connection_free(telumdb_connection_t* connection);

/* Query execution */
telumdb_error_t telumdb_execute(
    telumdb_connection_t* connection,
    const char* query,
    telumdb_result_t** result
);

telumdb_error_t telumdb_execute_params(
    telumdb_connection_t* connection,
    const char* query,
    const char* const* param_names,
    const void* const* param_values,
    const telumdb_data_type_t* param_types,
    int param_count,
    telumdb_result_t** result
);

/* Result handling */
telumdb_error_t telumdb_result_get_row_count(const telumdb_result_t* result, int64_t* row_count);
telumdb_error_t telumdb_result_get_column_count(const telumdb_result_t* result, int* column_count);
telumdb_error_t telumdb_result_get_column_name(const telumdb_result_t* result, int column, const char** name);
telumdb_error_t telumdb_result_get_column_type(const telumdb_result_t* result, int column, telumdb_data_type_t* type);
telumdb_error_t telumdb_result_get_affected_rows(const telumdb_result_t* result, int64_t* affected_rows);

telumdb_error_t telumdb_result_get_value(
    const telumdb_result_t* result,
    int row,
    int column,
    void** value,
    size_t* size
);

telumdb_error_t telumdb_result_get_int32(const telumdb_result_t* result, int row, int column, int32_t* value);
telumdb_error_t telumdb_result_get_int64(const telumdb_result_t* result, int row, int column, int64_t* value);
telumdb_error_t telumdb_result_get_float32(const telumdb_result_t* result, int row, int column, float* value);
telumdb_error_t telumdb_result_get_float64(const telumdb_result_t* result, int row, int column, double* value);
telumdb_error_t telumdb_result_get_string(const telumdb_result_t* result, int row, int column, const char** value);
telumdb_error_t telumdb_result_get_bool(const telumdb_result_t* result, int row, int column, bool* value);

telumdb_error_t telumdb_result_free(telumdb_result_t* result);

/* Tensor operations */
telumdb_error_t telumdb_create_tensor(
    telumdb_connection_t* connection,
    const char* name,
    const int64_t* shape,
    int shape_dims,
    telumdb_tensor_dtype_t dtype,
    telumdb_tensor_t** tensor
);

telumdb_error_t telumdb_get_tensor(
    telumdb_connection_t* connection,
    const char* name,
    telumdb_tensor_t** tensor
);

telumdb_error_t telumdb_tensor_get_name(const telumdb_tensor_t* tensor, const char** name);
telumdb_error_t telumdb_tensor_get_shape(const telumdb_tensor_t* tensor, const int64_t** shape, int* dims);
telumdb_error_t telumdb_tensor_get_dtype(const telumdb_tensor_t* tensor, telumdb_tensor_dtype_t* dtype);
telumdb_error_t telumdb_tensor_get_size(const telumdb_tensor_t* tensor, size_t* size);

telumdb_error_t telumdb_tensor_store_chunk(
    telumdb_tensor_t* tensor,
    const int64_t* start_indices,
    const int64_t* chunk_shape,
    const void* data,
    size_t data_size
);

telumdb_error_t telumdb_tensor_get_chunk(
    telumdb_tensor_t* tensor,
    const int64_t* start_indices,
    const int64_t* chunk_shape,
    void** data,
    size_t* data_size
);

telumdb_error_t telumdb_tensor_slice(
    telumdb_tensor_t* tensor,
    const int64_t* start,
    const int64_t* end,
    telumdb_tensor_t** result
);

telumdb_error_t telumdb_tensor_reshape(
    telumdb_tensor_t* tensor,
    const int64_t* new_shape,
    int new_dims
);

telumdb_error_t telumdb_tensor_add(
    const telumdb_tensor_t* a,
    const telumdb_tensor_t* b,
    telumdb_tensor_t** result
);

telumdb_error_t telumdb_tensor_multiply(
    const telumdb_tensor_t* a,
    const telumdb_tensor_t* b,
    telumdb_tensor_t** result
);

telumdb_error_t telumdb_tensor_cosine_similarity(
    const telumdb_tensor_t* a,
    const telumdb_tensor_t* b,
    float* similarity
);

telumdb_error_t telumdb_tensor_free(telumdb_tensor_t* tensor);

/* Utility functions */
const char* telumdb_error_string(telumdb_error_t error);
const char* telumdb_version(void);
telumdb_error_t telumdb_get_server_info(telumdb_connection_t* connection, char** info);

/* Batch operations */
typedef struct telumdb_batch telumdb_batch_t;

telumdb_error_t telumdb_batch_init(telumdb_batch_t** batch);
telumdb_error_t telumdb_batch_add_query(telumdb_batch_t* batch, const char* query);
telumdb_error_t telumdb_batch_execute(telumdb_connection_t* connection, telumdb_batch_t* batch, telumdb_result_t*** results, int* result_count);
telumdb_error_t telumdb_batch_free(telumdb_batch_t* batch);

/* Async operations (optional) */
typedef struct telumdb_future telumdb_future_t;

telumdb_error_t telumdb_execute_async(
    telumdb_connection_t* connection,
    const char* query,
    telumdb_future_t** future
);

telumdb_error_t telumdb_future_wait(telumdb_future_t* future, telumdb_result_t** result);
telumdb_error_t telumdb_future_free(telumdb_future_t* future);

#ifdef __cplusplus
}
#endif

#endif /* TELUMDB_H */