import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:user_flutter_v0_0_1/core/config/app.dart';

void main() {
  testWidgets('App shows home screen welcome text', (WidgetTester tester) async {
    await tester.pumpWidget(const ProviderScope(child: App()));
    await tester.pumpAndSettle();

    expect(find.byType(AppBar), findsOneWidget);
    expect(find.text('Home'), findsOneWidget);
    expect(find.text('Welcome to Trusioo User App!'), findsOneWidget);
    expect(find.text('开始使用'), findsOneWidget);
  });
}
