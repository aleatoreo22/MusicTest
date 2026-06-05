// Testes unitários da camada de API:
// - Models fazem parse correto do JSON do backend
// - ApiClient monta URLs/headers esperados e serializa errors

import 'dart:convert';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';

import 'package:music_unified/api/client.dart';
import 'package:music_unified/api/models.dart';

void main() {
  group('Models — parse', () {
    test('Track.fromJson lida com campos ausentes', () {
      final t = Track.fromJson({
        'id': 1,
        'title': 'Test',
        'artist': 'Artist',
        'duration': 180,
        'thumbnail': '',
        'preview_url': '',
        'source': 'spotify',
        'source_id': 'abc',
        'created_at': '2024-01-01T00:00:00Z',
      });
      expect(t.title, 'Test');
      expect(t.source, 'spotify');
      expect(t.duration, 180);
    });

    test('Track.fromJson usa defaults razoáveis', () {
      final t = Track.fromJson({});
      expect(t.title, '');
      expect(t.artist, 'Unknown');
      expect(t.duration, 0);
    });

    test('AuthResponse.fromJson', () {
      final r = AuthResponse.fromJson({
        'message': 'ok',
        'token': 'jwt.abc.def',
        'type': 'Bearer',
        'user_id': 42,
      });
      expect(r.token, 'jwt.abc.def');
      expect(r.userId, 42);
    });
  });

  group('ApiClient', () {
    test('login() POSTa no endpoint correto e parseia resposta', () async {
      final mock = MockClient((req) async {
        expect(req.method, 'POST');
        expect(req.url.path, '/api/v1/auth/login');
        expect(req.headers['Content-Type'], contains('application/json'));
        final body = jsonDecode(req.body) as Map<String, dynamic>;
        expect(body['email'], 'a@b.com');
        expect(body['password'], 'secret');

        return http.Response(
          jsonEncode({
            'message': 'ok',
            'token': 'tok-123',
            'type': 'Bearer',
            'user_id': 7,
          }),
          200,
          headers: {'content-type': 'application/json'},
        );
      });

      final api = ApiClient(baseUrl: 'http://x', client: mock);
      final resp = await api.login('a@b.com', 'secret');
      expect(resp.token, 'tok-123');
      expect(resp.userId, 7);
    });

    test('login() lança APIException com mensagem do backend em 4xx', () async {
      final mock = MockClient((req) async {
        return http.Response(
          jsonEncode({'error': 'Credenciais inválidas'}),
          401,
          headers: {'content-type': 'application/json'},
        );
      });

      final api = ApiClient(baseUrl: 'http://x', client: mock);
      expect(
        () => api.login('a@b.com', 'wrong'),
        throwsA(isA<APIException>()
            .having((e) => e.statusCode, 'statusCode', 401)
            .having((e) => e.message, 'message', 'Credenciais inválidas')),
      );
    });

    test('setToken injeta Authorization header', () async {
      String? receivedAuth;
      final mock = MockClient((req) async {
        receivedAuth = req.headers['Authorization'];
        return http.Response('[]', 200,
            headers: {'content-type': 'application/json'});
      });

      final api = ApiClient(baseUrl: 'http://x', client: mock);
      api.setToken('jwt-xyz');
      await api.searchSpotify('test');

      expect(receivedAuth, 'Bearer jwt-xyz');
    });

    test('getProviders retorna lista de strings', () async {
      final mock = MockClient((req) async {
        return http.Response(
          jsonEncode({'providers': ['spotify', 'soundcloud']}),
          200,
          headers: {'content-type': 'application/json'},
        );
      });

      final api = ApiClient(baseUrl: 'http://x', client: mock);
      final list = await api.getProviders();
      expect(list, ['spotify', 'soundcloud']);
    });
  });
}
