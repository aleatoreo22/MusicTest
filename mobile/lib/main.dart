import 'package:flutter/material.dart';

import 'api/auth_store.dart';
import 'api/client.dart';
import 'screens/home_screen.dart';
import 'screens/login_screen.dart';

void main() {
  runApp(const MusicUnifiedApp());
}

class MusicUnifiedApp extends StatelessWidget {
  const MusicUnifiedApp({super.key});

  // Base URL do backend. Para emulador Android, 10.0.2.2 mapeia para
  // o host (localhost da máquina de dev). Para Linux desktop, use 127.0.0.1.
  // Para dispositivo físico, substitua pelo IP da máquina na rede local.
  static const String _defaultBaseUrl = 'http://10.0.2.2:8080';

  @override
  Widget build(BuildContext context) {
    final api = ApiClient(baseUrl: _defaultBaseUrl);
    final auth = AuthStore();

    return MaterialApp(
      title: 'MusicUnified',
      debugShowCheckedModeBanner: false,
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(seedColor: Colors.deepPurple),
        useMaterial3: true,
      ),
      darkTheme: ThemeData(
        colorScheme: ColorScheme.fromSeed(
          seedColor: Colors.deepPurple,
          brightness: Brightness.dark,
        ),
        useMaterial3: true,
      ),
      home: _AuthGate(api: api, auth: auth),
    );
  }
}

/// _AuthGate decide a tela inicial com base no token persistido.
/// Se há token salvo, vai direto pra Home; senão, mostra Login.
class _AuthGate extends StatefulWidget {
  final ApiClient api;
  final AuthStore auth;

  const _AuthGate({required this.api, required this.auth});

  @override
  State<_AuthGate> createState() => _AuthGateState();
}

class _AuthGateState extends State<_AuthGate> {
  Widget? _initial;

  @override
  void initState() {
    super.initState();
    _bootstrap();
  }

  Future<void> _bootstrap() async {
    final token = await widget.auth.token();
    Widget screen;
    if (token != null && token.isNotEmpty) {
      widget.api.setToken(token);
      screen = HomeScreen(api: widget.api, auth: widget.auth);
    } else {
      screen = LoginScreen(api: widget.api, auth: widget.auth);
    }
    if (mounted) setState(() => _initial = screen);
  }

  @override
  Widget build(BuildContext context) {
    if (_initial == null) {
      return const Scaffold(
        body: Center(child: CircularProgressIndicator()),
      );
    }
    return _initial!;
  }
}
