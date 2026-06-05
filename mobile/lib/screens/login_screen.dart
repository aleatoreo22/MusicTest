import 'package:flutter/material.dart';

import '../api/auth_store.dart';
import '../api/client.dart';
import '../api/models.dart';
import 'home_screen.dart';

/// LoginScreen mostra um formulário de login/registro.
/// Após sucesso, navega para HomeScreen e remove a si mesma do stack.
class LoginScreen extends StatefulWidget {
  final ApiClient api;
  final AuthStore auth;

  const LoginScreen({super.key, required this.api, required this.auth});

  @override
  State<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends State<LoginScreen> {
  final _formKey = GlobalKey<FormState>();
  final _emailCtrl = TextEditingController();
  final _passwordCtrl = TextEditingController();
  bool _isRegister = false;
  bool _loading = false;
  String? _error;

  @override
  void dispose() {
    _emailCtrl.dispose();
    _passwordCtrl.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    if (!_formKey.currentState!.validate()) return;
    setState(() {
      _loading = true;
      _error = null;
    });

    try {
      final email = _emailCtrl.text.trim();
      final password = _passwordCtrl.text;
      final AuthResponse resp = _isRegister
          ? await widget.api.register(email, password)
          : await widget.api.login(email, password);

      await widget.auth.save(
        token: resp.token,
        userId: resp.userId,
        email: email,
      );
      widget.api.setToken(resp.token);

      if (!mounted) return;
      Navigator.of(context).pushReplacement(
        MaterialPageRoute(
          builder: (_) => HomeScreen(api: widget.api, auth: widget.auth),
        ),
      );
    } on APIException catch (e) {
      setState(() => _error = e.message);
    } catch (e) {
      setState(() => _error = 'Erro de conexão: $e');
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: SafeArea(
        child: Center(
          child: SingleChildScrollView(
            padding: const EdgeInsets.all(24),
            child: ConstrainedBox(
              constraints: const BoxConstraints(maxWidth: 400),
              child: Form(
                key: _formKey,
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    const Icon(Icons.music_note, size: 64),
                    const SizedBox(height: 16),
                    Text(
                      'MusicUnified',
                      textAlign: TextAlign.center,
                      style: Theme.of(context).textTheme.headlineSmall,
                    ),
                    const SizedBox(height: 32),
                    TextFormField(
                      controller: _emailCtrl,
                      decoration: const InputDecoration(
                        labelText: 'E-mail',
                        border: OutlineInputBorder(),
                        prefixIcon: Icon(Icons.email),
                      ),
                      keyboardType: TextInputType.emailAddress,
                      autocorrect: false,
                      validator: (v) {
                        if (v == null || v.trim().isEmpty) {
                          return 'Informe o e-mail';
                        }
                        if (!v.contains('@')) return 'E-mail inválido';
                        return null;
                      },
                    ),
                    const SizedBox(height: 16),
                    TextFormField(
                      controller: _passwordCtrl,
                      decoration: const InputDecoration(
                        labelText: 'Senha',
                        border: OutlineInputBorder(),
                        prefixIcon: Icon(Icons.lock),
                      ),
                      obscureText: true,
                      validator: (v) {
                        if (v == null || v.isEmpty) {
                          return 'Informe a senha';
                        }
                        if (_isRegister && v.length < 6) {
                          return 'Mínimo 6 caracteres';
                        }
                        return null;
                      },
                    ),
                    if (_error != null) ...[
                      const SizedBox(height: 16),
                      Container(
                        padding: const EdgeInsets.all(12),
                        decoration: BoxDecoration(
                          color: Colors.red.shade50,
                          border: Border.all(color: Colors.red.shade200),
                          borderRadius: BorderRadius.circular(8),
                        ),
                        child: Text(
                          _error!,
                          style: TextStyle(color: Colors.red.shade900),
                        ),
                      ),
                    ],
                    const SizedBox(height: 24),
                    FilledButton(
                      onPressed: _loading ? null : _submit,
                      child: _loading
                          ? const SizedBox(
                              height: 20,
                              width: 20,
                              child: CircularProgressIndicator(
                                strokeWidth: 2,
                                color: Colors.white,
                              ),
                            )
                          : Text(_isRegister ? 'Criar conta' : 'Entrar'),
                    ),
                    const SizedBox(height: 12),
                    TextButton(
                      onPressed: _loading
                          ? null
                          : () => setState(() => _isRegister = !_isRegister),
                      child: Text(_isRegister
                          ? 'Já tem conta? Entrar'
                          : 'Não tem conta? Registre-se'),
                    ),
                  ],
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }
}
