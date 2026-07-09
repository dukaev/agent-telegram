//go:build darwin && cgo

package sessionstore

/*
#cgo LDFLAGS: -framework Security -framework CoreFoundation
#include <CoreFoundation/CoreFoundation.h>
#include <Security/Security.h>
#include <stdint.h>
#include <stdlib.h>

static CFStringRef ATString(const char *value, CFIndex length) {
  return CFStringCreateWithBytes(
      kCFAllocatorDefault, (const UInt8 *)value, length,
      kCFStringEncodingUTF8, false);
}

static CFMutableDictionaryRef ATQuery(
    const char *service, CFIndex serviceLength,
    const char *account, CFIndex accountLength) {
  CFStringRef serviceValue = ATString(service, serviceLength);
  CFStringRef accountValue = ATString(account, accountLength);
  if (serviceValue == NULL || accountValue == NULL) {
    if (serviceValue != NULL) CFRelease(serviceValue);
    if (accountValue != NULL) CFRelease(accountValue);
    return NULL;
  }
  CFMutableDictionaryRef query = CFDictionaryCreateMutable(
      kCFAllocatorDefault, 0,
      &kCFTypeDictionaryKeyCallBacks, &kCFTypeDictionaryValueCallBacks);
  if (query != NULL) {
    CFDictionarySetValue(query, kSecClass, kSecClassGenericPassword);
    CFDictionarySetValue(query, kSecAttrService, serviceValue);
    CFDictionarySetValue(query, kSecAttrAccount, accountValue);
  }
  CFRelease(serviceValue);
  CFRelease(accountValue);
  return query;
}

static OSStatus ATKeychainLoad(
    const char *service, CFIndex serviceLength,
    const char *account, CFIndex accountLength,
    UInt32 *dataLength, void **data) {
  CFMutableDictionaryRef query = ATQuery(
      service, serviceLength, account, accountLength);
  if (query == NULL) return errSecAllocate;
  CFDictionarySetValue(query, kSecReturnData, kCFBooleanTrue);
  CFDictionarySetValue(query, kSecMatchLimit, kSecMatchLimitOne);

  CFTypeRef result = NULL;
  OSStatus status = SecItemCopyMatching(query, &result);
  CFRelease(query);
  if (status != errSecSuccess) return status;

  CFDataRef value = (CFDataRef)result;
  CFIndex length = CFDataGetLength(value);
  if (length < 0 || (uint64_t)length > UINT32_MAX) {
    CFRelease(result);
    return errSecDecode;
  }
  void *buffer = malloc((size_t)length);
  if (buffer == NULL && length > 0) {
    CFRelease(result);
    return errSecAllocate;
  }
  if (length > 0) {
    CFDataGetBytes(value, CFRangeMake(0, length), (UInt8 *)buffer);
  }
  CFRelease(result);
  *dataLength = (UInt32)length;
  *data = buffer;
  return errSecSuccess;
}

static OSStatus ATKeychainSave(
    const char *service, CFIndex serviceLength,
    const char *account, CFIndex accountLength,
    const void *data, CFIndex dataLength) {
  CFMutableDictionaryRef query = ATQuery(
      service, serviceLength, account, accountLength);
  if (query == NULL) return errSecAllocate;
  CFDataRef value = CFDataCreate(
      kCFAllocatorDefault, (const UInt8 *)data, dataLength);
  if (value == NULL) {
    CFRelease(query);
    return errSecAllocate;
  }

  const void *keys[] = { kSecValueData };
  const void *values[] = { value };
  CFDictionaryRef attributes = CFDictionaryCreate(
      kCFAllocatorDefault, keys, values, 1,
      &kCFTypeDictionaryKeyCallBacks, &kCFTypeDictionaryValueCallBacks);
  OSStatus status = SecItemUpdate(query, attributes);
  if (status == errSecItemNotFound) {
    CFDictionarySetValue(query, kSecValueData, value);
    CFDictionarySetValue(query, kSecAttrAccessible, kSecAttrAccessibleAfterFirstUnlock);
    status = SecItemAdd(query, NULL);
  }
  CFRelease(attributes);
  CFRelease(value);
  CFRelease(query);
  return status;
}

static OSStatus ATKeychainDelete(
    const char *service, CFIndex serviceLength,
    const char *account, CFIndex accountLength) {
  CFMutableDictionaryRef query = ATQuery(
      service, serviceLength, account, accountLength);
  if (query == NULL) return errSecAllocate;
  OSStatus status = SecItemDelete(query);
  CFRelease(query);
  if (status == errSecItemNotFound) return errSecSuccess;
  return status;
}

static void ATKeychainFree(void *data) {
  if (data != NULL) free(data);
}

static char *ATKeychainError(OSStatus status) {
  CFStringRef message = SecCopyErrorMessageString(status, NULL);
  if (message == NULL) {
    return NULL;
  }
  CFIndex maxLength = CFStringGetMaximumSizeForEncoding(
      CFStringGetLength(message), kCFStringEncodingUTF8) + 1;
  char *buffer = (char *)malloc((size_t)maxLength);
  if (buffer == NULL) {
    CFRelease(message);
    return NULL;
  }
  if (!CFStringGetCString(message, buffer, maxLength, kCFStringEncodingUTF8)) {
    free(buffer);
    buffer = NULL;
  }
  CFRelease(message);
  return buffer;
}
*/
import "C"

import (
	"context"
	"fmt"
	"math"
	"runtime"
	"unsafe"
)

const (
	KeychainProvider = "keychain"
	keychainService  = "io.agent-telegram.session"
	errSecSuccess    = 0
	errSecNotFound   = -25300
)

type keychainStore struct{}

func init() {
	RegisterProvider(KeychainProvider, func() (Store, error) { return &keychainStore{}, nil })
}

func (s *keychainStore) Provider() string { return KeychainProvider }
func (s *keychainStore) Persistent() bool { return true }

func (s *keychainStore) Load(ctx context.Context, profile string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	service := C.CString(keychainService)
	account := C.CString(profile)
	defer C.free(unsafe.Pointer(service))
	defer C.free(unsafe.Pointer(account))

	var length C.UInt32
	var data unsafe.Pointer
	status := C.ATKeychainLoad(
		service, C.CFIndex(len(keychainService)),
		account, C.CFIndex(len(profile)),
		&length, &data,
	)
	if int32(status) == errSecNotFound {
		return nil, ErrNotFound
	}
	if int32(status) != errSecSuccess {
		return nil, keychainError("load", status)
	}
	defer C.ATKeychainFree(data)
	if length == 0 || data == nil {
		return nil, ErrNotFound
	}
	return C.GoBytes(data, C.int(length)), nil
}

func (s *keychainStore) Save(ctx context.Context, profile string, data []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if len(data) == 0 {
		return fmt.Errorf("session is empty")
	}
	if uint64(len(data)) > math.MaxUint32 {
		return fmt.Errorf("session is too large")
	}
	service := C.CString(keychainService)
	account := C.CString(profile)
	defer C.free(unsafe.Pointer(service))
	defer C.free(unsafe.Pointer(account))

	status := C.ATKeychainSave(
		service, C.CFIndex(len(keychainService)),
		account, C.CFIndex(len(profile)),
		unsafe.Pointer(&data[0]), C.CFIndex(len(data)),
	)
	runtime.KeepAlive(data)
	if int32(status) != errSecSuccess {
		return keychainError("save", status)
	}
	return nil
}

func (s *keychainStore) Delete(ctx context.Context, profile string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	service := C.CString(keychainService)
	account := C.CString(profile)
	defer C.free(unsafe.Pointer(service))
	defer C.free(unsafe.Pointer(account))

	status := C.ATKeychainDelete(
		service, C.CFIndex(len(keychainService)),
		account, C.CFIndex(len(profile)),
	)
	if int32(status) != errSecSuccess {
		return keychainError("delete", status)
	}
	return nil
}

func keychainError(action string, status C.OSStatus) error {
	message := C.ATKeychainError(status)
	if message == nil {
		return fmt.Errorf("macOS Keychain %s failed (OSStatus %d)", action, int32(status))
	}
	defer C.free(unsafe.Pointer(message))
	return fmt.Errorf("macOS Keychain %s failed: %s (OSStatus %d)", action, C.GoString(message), int32(status))
}
