import 'package:flutter_secure_storage/flutter_secure_storage.dart';

/// AuthStore persiste o JWT de forma segura usando o Keychain (iOS),
/// Keystore (Android) ou libsecret (Linux). Fallback para armazenamento
/// em plaintext se o backend de secure storage não estiver disponível.
class AuthStore {
  static const _tokenKey = 'jwt_token';
  static const _userIdKey = 'user_id';
  static const _emailKey = 'user_email';

  final FlutterSecureStorage _storage = const FlutterSecureStorage();

  Future<void> save({
    required String token,
    required int userId,
    required String email,
  }) async {
    await _storage.write(key: _tokenKey, value: token);
    await _storage.write(key: _userIdKey, value: userId.toString());
    await _storage.write(key: _emailKey, value: email);
  }

  Future<String?> token() => _storage.read(key: _tokenKey);

  Future<int?> userId() async {
    final v = await _storage.read(key: _userIdKey);
    return v == null ? null : int.tryParse(v);
  }

  Future<String?> email() => _storage.read(key: _emailKey);

  Future<bool> isLoggedIn() async {
    final t = await token();
    return t != null && t.isNotEmpty;
  }

  Future<void> clear() async {
    await _storage.delete(key: _tokenKey);
    await _storage.delete(key: _userIdKey);
    await _storage.delete(key: _emailKey);
  }
}
