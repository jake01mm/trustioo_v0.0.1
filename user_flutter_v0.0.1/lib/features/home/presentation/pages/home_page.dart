import 'package:flutter/material.dart';

class HomePage extends StatelessWidget {
  const HomePage({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Home')),
      body: Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Text('Welcome to Trusioo User App!'),
            const SizedBox(height: 16),
            ElevatedButton(
              onPressed: () {},
              child: const Text('开始使用'),
            )
          ],
        ),
      ),
    );
  }
}
