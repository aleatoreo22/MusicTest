import 'package:flutter/material.dart';

import '../api/auth_store.dart';
import '../api/client.dart';
import 'login_screen.dart';

/// HomeScreen é o placeholder pós-login. Será expandida nas próximas
/// iterações (busca, playlists, player). Por enquanto: lista providers
/// habilitados, mostra dados do usuário logado, e tem botão de logout.
class HomeScreen extends StatefulWidget {
  final ApiClient api;
  final AuthStore auth;

  const HomeScreen({super.key, required this.api, required this.auth});

  @override
  State<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends State<HomeScreen> {
  List<String> _providers = [];
  bool _loading = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final p = await widget.api.getProviders();
      setState(() {
        _providers = p;
        _loading = false;
      });
    } on APIException catch (e) {
      setState(() {
        _error = e.message;
        _loading = false;
      });
    } catch (e) {
      setState(() {
        _error = 'Erro: $e';
        _loading = false;
      });
    }
  }

  Future<void> _logout() async {
    await widget.auth.clear();
    widget.api.setToken(null);
    if (!mounted) return;
    Navigator.of(context).pushReplacement(
      MaterialPageRoute(
        builder: (_) => LoginScreen(api: widget.api, auth: widget.auth),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('MusicUnified'),
        actions: [
          IconButton(
            icon: const Icon(Icons.logout),
            onPressed: _logout,
            tooltip: 'Sair',
          ),
        ],
      ),
      body: RefreshIndicator(
        onRefresh: _load,
        child: ListView(
          padding: const EdgeInsets.all(16),
          children: [
            Card(
              child: Padding(
                padding: const EdgeInsets.all(16),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      'Conectado',
                      style: Theme.of(context).textTheme.titleMedium,
                    ),
                    const SizedBox(height: 8),
                    FutureBuilder<String?>(
                      future: widget.auth.email(),
                      builder: (ctx, snap) => Text(
                        snap.data ?? '...',
                        style: Theme.of(context).textTheme.bodyLarge,
                      ),
                    ),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 16),
            Text(
              'Provedores habilitados',
              style: Theme.of(context).textTheme.titleMedium,
            ),
            const SizedBox(height: 8),
            if (_loading)
              const Center(child: CircularProgressIndicator())
            else if (_error != null)
              Card(
                color: Colors.red.shade50,
                child: Padding(
                  padding: const EdgeInsets.all(16),
                  child: Text(_error!),
                ),
              )
            else if (_providers.isEmpty)
              const Card(
                child: Padding(
                  padding: EdgeInsets.all(16),
                  child: Text('Nenhum provedor configurado no backend.'),
                ),
              )
            else
              ..._providers.map(
                (p) => Card(
                  child: ListTile(
                    leading: Icon(_iconFor(p)),
                    title: Text(p),
                    trailing: const Icon(Icons.chevron_right),
                    onTap: () {
                      // TODO: navegar para busca deste provider
                    },
                  ),
                ),
              ),
          ],
        ),
      ),
    );
  }

  IconData _iconFor(String provider) {
    switch (provider) {
      case 'spotify':
        return Icons.music_note;
      case 'soundcloud':
        return Icons.cloud;
      case 'youtube':
        return Icons.play_circle_outline;
      default:
        return Icons.audio_file;
    }
  }
}
