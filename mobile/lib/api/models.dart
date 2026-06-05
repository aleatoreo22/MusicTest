// Models que espelham os JSONs do backend Go.
// Mantidos no mesmo arquivo por simplicidade inicial; podem ser
// quebrados por domínio (user, track, playlist) conforme crescerem.

class User {
  final int id;
  final String email;
  final String token;
  final DateTime createdAt;

  const User({
    required this.id,
    required this.email,
    required this.token,
    required this.createdAt,
  });

  factory User.fromJson(Map<String, dynamic> json) => User(
        id: json['id'] as int,
        email: json['email'] as String,
        token: (json['token'] as String?) ?? '',
        createdAt: DateTime.parse(json['created_at'] as String),
      );
}

class Track {
  final int id;
  final String title;
  final String artist;
  final int duration; // segundos
  final String thumbnail;
  final String previewUrl;
  final String source; // "spotify" | "soundcloud" | "youtube" | "local"
  final String sourceId;
  final DateTime createdAt;

  const Track({
    required this.id,
    required this.title,
    required this.artist,
    required this.duration,
    required this.thumbnail,
    required this.previewUrl,
    required this.source,
    required this.sourceId,
    required this.createdAt,
  });

  factory Track.fromJson(Map<String, dynamic> json) => Track(
        id: (json['id'] as num?)?.toInt() ?? 0,
        title: (json['title'] as String?) ?? '',
        artist: (json['artist'] as String?) ?? 'Unknown',
        duration: (json['duration'] as num?)?.toInt() ?? 0,
        thumbnail: (json['thumbnail'] as String?) ?? '',
        previewUrl: (json['preview_url'] as String?) ?? '',
        source: (json['source'] as String?) ?? '',
        sourceId: (json['source_id'] as String?) ?? '',
        createdAt: json['created_at'] != null
            ? DateTime.parse(json['created_at'] as String)
            : DateTime.now(),
      );
}

class Playlist {
  final int id;
  final String name;
  final List<Track> tracks;
  final DateTime createdAt;

  const Playlist({
    required this.id,
    required this.name,
    required this.tracks,
    required this.createdAt,
  });

  factory Playlist.fromJson(Map<String, dynamic> json) => Playlist(
        id: (json['id'] as num?)?.toInt() ?? 0,
        name: (json['name'] as String?) ?? '',
        tracks: ((json['tracks'] as List<dynamic>?) ?? [])
            .map((e) => Track.fromJson(e as Map<String, dynamic>))
            .toList(),
        createdAt: json['created_at'] != null
            ? DateTime.parse(json['created_at'] as String)
            : DateTime.now(),
      );
}

class AuthRequest {
  final String email;
  final String password;

  const AuthRequest({required this.email, required this.password});

  Map<String, dynamic> toJson() =>
      {'email': email, 'password': password};
}

class AuthResponse {
  final String message;
  final String token;
  final String type; // "Bearer"
  final int userId;

  const AuthResponse({
    required this.message,
    required this.token,
    required this.type,
    required this.userId,
  });

  factory AuthResponse.fromJson(Map<String, dynamic> json) => AuthResponse(
        message: (json['message'] as String?) ?? '',
        token: json['token'] as String,
        type: (json['type'] as String?) ?? 'Bearer',
        userId: (json['user_id'] as num?)?.toInt() ?? 0,
      );
}

class StreamResponse {
  final String url;
  final int expiresAt; // epoch seconds
  final String format; // "hls" | "audio"

  const StreamResponse({
    required this.url,
    required this.expiresAt,
    required this.format,
  });

  factory StreamResponse.fromJson(Map<String, dynamic> json) => StreamResponse(
        url: json['url'] as String,
        expiresAt: (json['expires_at'] as num).toInt(),
        format: (json['format'] as String?) ?? '',
      );
}

class PlaybackSnapshot {
  final bool isPlaying;
  final int progressMs;
  final int timestamp; // epoch ms
  final String uri;
  final String itemId;
  final int itemDurationMs;

  const PlaybackSnapshot({
    required this.isPlaying,
    required this.progressMs,
    required this.timestamp,
    required this.uri,
    required this.itemId,
    required this.itemDurationMs,
  });

  factory PlaybackSnapshot.fromJson(Map<String, dynamic> json) =>
      PlaybackSnapshot(
        isPlaying: (json['is_playing'] as bool?) ?? false,
        progressMs: (json['progress_ms'] as num?)?.toInt() ?? 0,
        timestamp: (json['timestamp'] as num?)?.toInt() ?? 0,
        uri: (json['uri'] as String?) ?? '',
        itemId: (json['item_id'] as String?) ?? '',
        itemDurationMs: (json['item_duration_ms'] as num?)?.toInt() ?? 0,
      );
}
