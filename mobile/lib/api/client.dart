import 'dart:async';
import 'dart:convert';

import 'package:http/http.dart' as http;

import 'models.dart';

/// APIException é a exceção de domínio para erros do backend.
/// Carrega o status HTTP e a mensagem retornada pelo servidor
/// (que vem em português do handler Go).
class APIException implements Exception {
  final int statusCode;
  final String message;

  const APIException(this.statusCode, this.message);

  @override
  String toString() => 'APIException($statusCode): $message';
}

/// ApiClient é o espelho em Dart do `cmd/client/api/client.go` Go.
/// Todas as chamadas HTTP vivem aqui; as telas consomem esta classe.
class ApiClient {
  final String baseUrl;
  String? _token;
  final http.Client _http;

  ApiClient({required this.baseUrl, http.Client? client})
      : _http = client ?? http.Client();

  void setToken(String? token) {
    _token = token;
  }

  String? get token => _token;

  Map<String, String> _headers({bool json = true}) {
    final h = <String, String>{};
    if (json) h['Content-Type'] = 'application/json';
    if (_token != null && _token!.isNotEmpty) {
      h['Authorization'] = 'Bearer $_token';
    }
    return h;
  }

  Future<dynamic> _request(
    String method,
    String endpoint, {
    Object? body,
    Map<String, String>? query,
  }) async {
    final uri = Uri.parse('$baseUrl$endpoint').replace(
      queryParameters: query,
    );

    final req = http.Request(method, uri);
    req.headers.addAll(_headers(json: body != null));
    if (body != null) {
      req.body = jsonEncode(body);
    }

    final streamed = await _http.send(req).timeout(
          const Duration(seconds: 15),
        );
    final resp = await http.Response.fromStream(streamed);

    if (resp.statusCode >= 400) {
      // Tenta extrair a mensagem {"error": "..."} que os handlers Go retornam
      String msg = resp.body;
      try {
        final parsed = jsonDecode(resp.body);
        if (parsed is Map && parsed['error'] is String) {
          msg = parsed['error'] as String;
        }
      } catch (_) {}
      throw APIException(resp.statusCode, msg);
    }

    if (resp.body.isEmpty) return null;
    return jsonDecode(resp.body);
  }

  // ===== Auth =====

  Future<AuthResponse> register(String email, String password) async {
    final data = await _request('POST', '/api/v1/auth/register',
        body: AuthRequest(email: email, password: password).toJson());
    return AuthResponse.fromJson(data as Map<String, dynamic>);
  }

  Future<AuthResponse> login(String email, String password) async {
    final data = await _request('POST', '/api/v1/auth/login',
        body: AuthRequest(email: email, password: password).toJson());
    return AuthResponse.fromJson(data as Map<String, dynamic>);
  }

  Future<Map<String, dynamic>> me() async {
    final data = await _request('GET', '/api/v1/me/profile');
    return data as Map<String, dynamic>;
  }

  // ===== Providers =====

  Future<List<String>> getProviders() async {
    final data = await _request('GET', '/api/v1/providers');
    return (data['providers'] as List<dynamic>).cast<String>();
  }

  // ===== Search =====

  Future<List<Track>> searchSpotify(String query) async {
    final data = await _request('GET', '/api/v1/search/spotify',
        query: {'q': query});
    return (data as List<dynamic>)
        .map((e) => Track.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<List<Track>> searchSoundCloud(String query,
      {int limit = 20}) async {
    final data = await _request('GET', '/api/v1/search/soundcloud',
        query: {'q': query, 'limit': '$limit'});
    return (data as List<dynamic>)
        .map((e) => Track.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<List<Track>> searchYouTube(String query, {int limit = 20}) async {
    final data = await _request('GET', '/api/v1/search/youtube',
        query: {'q': query, 'limit': '$limit'});
    return (data as List<dynamic>)
        .map((e) => Track.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  // ===== Stream =====

  /// getStreamURL resolve a URL reproduzível para um track.
  /// Source deve ser "soundcloud" ou "youtube" (Spotify usa Connect).
  Future<StreamResponse> getStreamURL(String source, String sourceId) async {
    final endpoint = '/api/v1/stream/$source/$sourceId';
    final data = await _request('GET', endpoint);
    return StreamResponse.fromJson(data as Map<String, dynamic>);
  }

  // ===== Playlists =====

  Future<List<Playlist>> getPlaylists() async {
    final data = await _request('GET', '/api/v1/playlists');
    return (data as List<dynamic>)
        .map((e) => Playlist.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<Playlist> createPlaylist(String name) async {
    final data = await _request('POST', '/api/v1/playlists',
        body: {'name': name});
    return Playlist.fromJson(data as Map<String, dynamic>);
  }

  Future<void> addTrackToPlaylist(int playlistId, Track track) async {
    await _request('POST', '/api/v1/playlists/$playlistId/tracks', body: {
      'title': track.title,
      'artist': track.artist,
      'duration': track.duration,
      'thumbnail': track.thumbnail,
      'source': track.source,
      'source_id': track.sourceId,
      'preview_url': track.previewUrl,
    });
  }

  // ===== Player (Spotify Connect) =====

  Future<PlaybackSnapshot> currentlyPlaying() async {
    final data = await _request('GET', '/api/v1/player/currently-playing');
    return PlaybackSnapshot.fromJson(data as Map<String, dynamic>);
  }

  Future<void> playTrack({String? uri, String? deviceId}) async {
    await _request('POST', '/api/v1/player/play', body: {
      if (uri != null) 'uri': uri,
      if (deviceId != null) 'device_id': deviceId,
    });
  }

  Future<void> pause({bool cleanup = false}) async {
    final query = cleanup ? {'cleanup': '1'} : null;
    await _request('POST', '/api/v1/player/pause', query: query);
  }

  Future<void> resume() async {
    await _request('POST', '/api/v1/player/play');
  }

  Future<void> next() async {
    await _request('POST', '/api/v1/player/next');
  }

  Future<void> previous() async {
    await _request('POST', '/api/v1/player/previous');
  }

  Future<void> seek(int positionMs, {String? deviceId}) async {
    await _request('POST', '/api/v1/player/seek', body: {
      'position_ms': positionMs,
      if (deviceId != null) 'device_id': deviceId,
    });
  }

  // ===== Spotify linking =====

  Future<bool> getSpotifyStatus() async {
    final data = await _request('GET', '/api/v1/auth/spotify/status');
    return (data as Map<String, dynamic>)['linked'] as bool? ?? false;
  }

  Future<String> getSpotifyAuthURL() async {
    final data = await _request('GET', '/api/v1/auth/spotify');
    return (data as Map<String, dynamic>)['url'] as String;
  }

  void close() => _http.close();
}
