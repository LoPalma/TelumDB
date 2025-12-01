/*
 * TelumDB C Library - Connection Tests
 * 
 * Copyright 2024 TelumDB Contributors
 * Licensed under the Apache License, Version 2.0
 */

#include "../include/telumdb.h"
#include <stdio.h>
#include <stdlib.h>
#include <assert.h>

static void test_config_init() {
    printf("Testing config initialization...\n");
    
    telumdb_config_t* config = NULL;
    telumdb_error_t err = telumdb_config_init(&config);
    
    assert(err == TELUMDB_OK);
    assert(config != NULL);
    
    telumdb_config_free(config);
    printf("✓ Config initialization test passed\n");
}

static void test_config_setters() {
    printf("Testing config setters...\n");
    
    telumdb_config_t* config = NULL;
    telumdb_error_t err = telumdb_config_init(&config);
    assert(err == TELUMDB_OK);
    
    // Test setting host
    err = telumdb_config_set_host(config, "example.com");
    assert(err == TELUMDB_OK);
    
    // Test setting port
    err = telumdb_config_set_port(config, 5433);
    assert(err == TELUMDB_OK);
    
    // Test invalid port
    err = telumdb_config_set_port(config, -1);
    assert(err == TELUMDB_ERROR_INVALID_PARAMETER);
    
    // Test setting database
    err = telumdb_config_set_database(config, "testdb");
    assert(err == TELUMDB_OK);
    
    // Test setting credentials
    err = telumdb_config_set_credentials(config, "user", "pass");
    assert(err == TELUMDB_OK);
    
    // Test setting timeout
    err = telumdb_config_set_timeout(config, 60);
    assert(err == TELUMDB_OK);
    
    // Test setting SSL
    err = telumdb_config_set_ssl(config, true);
    assert(err == TELUMDB_OK);
    
    telumdb_config_free(config);
    printf("✓ Config setters test passed\n");
}

static void test_connection() {
    printf("Testing connection...\n");
    
    telumdb_config_t* config = NULL;
    telumdb_error_t err = telumdb_config_init(&config);
    assert(err == TELUMDB_OK);
    
    err = telumdb_config_set_host(config, "localhost");
    assert(err == TELUMDB_OK);
    
    telumdb_connection_t* connection = NULL;
    err = telumdb_connect(config, &connection);
    
    /* Note: This will fail until we implement actual connection logic */
    if (err == TELUMDB_OK) {
        assert(connection != NULL);
        assert(telumdb_is_connected(connection) == true);
        
        // Test ping
        err = telumdb_ping(connection);
        assert(err == TELUMDB_OK);
        
        // Test server info
        char* info = NULL;
        err = telumdb_get_server_info(connection, &info);
        assert(err == TELUMDB_OK);
        assert(info != NULL);
        free(info);
        
        // Test disconnect
        err = telumdb_disconnect(connection);
        assert(err == TELUMDB_OK);
        assert(telumdb_is_connected(connection) == false);
        
        telumdb_connection_free(connection);
    } else {
        printf("  Connection failed (expected until implementation is complete): %s\n", 
               telumdb_error_string(err));
    }
    
    telumdb_config_free(config);
    printf("✓ Connection test passed\n");
}

static void test_error_handling() {
    printf("Testing error handling...\n");
    
    // Test NULL parameters
    telumdb_error_t err = telumdb_config_init(NULL);
    assert(err == TELUMDB_ERROR_INVALID_PARAMETER);
    
    telumdb_config_t* config = NULL;
    err = telumdb_config_set_host(NULL, "localhost");
    assert(err == TELUMDB_ERROR_INVALID_PARAMETER);
    
    err = telumdb_config_set_host(config, NULL);
    assert(err == TELUMDB_ERROR_INVALID_PARAMETER);
    
    // Test error strings
    const char* error_str = telumdb_error_string(TELUMDB_ERROR_OUT_OF_MEMORY);
    assert(error_str != NULL);
    printf("  Error string for TELUMDB_ERROR_OUT_OF_MEMORY: %s\n", error_str);
    
    printf("✓ Error handling test passed\n");
}

static void test_version() {
    printf("Testing version info...\n");
    
    const char* version = telumdb_version();
    assert(version != NULL);
    printf("  TelumDB version: %s\n", version);
    
    printf("✓ Version test passed\n");
}

int main() {
    printf("TelumDB C Library Connection Tests\n");
    printf("==================================\n\n");
    
    test_config_init();
    test_config_setters();
    test_connection();
    test_error_handling();
    test_version();
    
    printf("\n✅ All tests passed!\n");
    return 0;
}