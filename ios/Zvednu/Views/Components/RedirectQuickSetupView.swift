import SwiftUI

struct RedirectQuickSetupView: View {
    let karenNumber: String
    @State private var noAnswerTime: Int = RedirectCodes.defaultNoAnswerTime
    @State private var expandedType: RedirectType?

    private var cleanKarenNumber: String {
        karenNumber.replacingOccurrences(of: " ", with: "")
    }

    var body: some View {
        ScrollView {
            VStack(spacing: 12) {
                // Instructions
                HStack(spacing: 8) {
                    Image(systemName: "info.circle.fill")
                        .foregroundStyle(.blue)
                    Text("Přesměrování se nastavuje vytočením speciálního kódu na telefonu.")
                        .font(.caption)
                }
                .padding()
                .background(Color.blue.opacity(0.1))
                .clipShape(RoundedRectangle(cornerRadius: 12))

                // Warning about clearing existing redirects
                HStack(spacing: 8) {
                    Image(systemName: "exclamationmark.triangle.fill")
                        .foregroundStyle(.orange)
                    VStack(alignment: .leading, spacing: 4) {
                        Text("Zobrazuje se chyba?")
                            .font(.caption)
                            .fontWeight(.medium)
                        Text("Nejdřív zruš stávající přesměrování pomocí \"Zrušit přesměrování\" níže, nebo vytočením kódu ##002#.")
                            .font(.caption2)
                    }
                }
                .padding()
                .background(Color.orange.opacity(0.1))
                .clipShape(RoundedRectangle(cornerRadius: 12))

                // Clear all button
                clearAllSection

                // Redirect type sections
                ForEach(RedirectType.allCases) { type in
                    redirectSection(for: type)
                }
            }
            .padding()
        }
    }

    private var clearAllSection: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text("Zrušit všechna přesměrování")
                .font(.subheadline)
                .fontWeight(.medium)

            HStack {
                Text(RedirectCodes.clearAllRedirectsCode)
                    .font(.system(.caption, design: .monospaced))
                    .foregroundStyle(Color.accentColor)

                Spacer()

                if let url = URL(string: "tel:\(RedirectCodes.clearAllRedirectsCode)") {
                    Link(destination: url) {
                        Text("Vytočit")
                            .font(.caption2)
                            .padding(.horizontal, 10)
                            .padding(.vertical, 6)
                            .background(Color.orange)
                            .foregroundStyle(.white)
                            .clipShape(Capsule())
                    }
                }
            }
        }
        .padding()
        .background(Color(.systemGray6))
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }

    private func redirectSection(for type: RedirectType) -> some View {
        let isExpanded = expandedType == type
        let dialCode = type == .noAnswer
            ? RedirectCodes.getDialCode(type: type, phoneNumber: cleanKarenNumber, time: noAnswerTime)
            : RedirectCodes.getDialCode(type: type, phoneNumber: cleanKarenNumber)

        return VStack(alignment: .leading, spacing: 0) {
            // Header (always visible)
            Button {
                withAnimation {
                    expandedType = isExpanded ? nil : type
                }
            } label: {
                HStack {
                    VStack(alignment: .leading, spacing: 2) {
                        Text(type.label)
                            .font(.subheadline)
                            .fontWeight(.medium)
                            .foregroundStyle(.primary)

                        Text(type.description(time: noAnswerTime))
                            .font(.caption2)
                            .foregroundStyle(.secondary)
                            .lineLimit(isExpanded ? nil : 1)
                    }

                    Spacer()

                    Image(systemName: isExpanded ? "chevron.up" : "chevron.down")
                        .font(.caption)
                        .foregroundStyle(.secondary)
                }
                .padding()
            }

            // Expanded content
            if isExpanded {
                VStack(alignment: .leading, spacing: 12) {
                    Divider()

                    // Timing selector for noAnswer
                    if type == .noAnswer {
                        VStack(alignment: .leading, spacing: 8) {
                            Text("Po kolika sekundách přesměrovat?")
                                .font(.caption)
                                .fontWeight(.medium)

                            ScrollView(.horizontal, showsIndicators: false) {
                                HStack(spacing: 8) {
                                    ForEach(RedirectCodes.noAnswerTimeOptions, id: \.self) { time in
                                        Button {
                                            noAnswerTime = time
                                        } label: {
                                            Text("\(time)s")
                                                .font(.caption)
                                                .padding(.horizontal, 12)
                                                .padding(.vertical, 6)
                                                .background(
                                                    noAnswerTime == time
                                                        ? Color.accentColor
                                                        : Color(.systemGray5)
                                                )
                                                .foregroundStyle(
                                                    noAnswerTime == time
                                                        ? .white
                                                        : .primary
                                                )
                                                .clipShape(Capsule())
                                        }
                                    }
                                }
                            }
                        }
                    }

                    // Activation code
                    VStack(alignment: .leading, spacing: 4) {
                        Text("Aktivační kód")
                            .font(.caption)
                            .foregroundStyle(.secondary)

                        HStack {
                            Text(dialCode)
                                .font(.system(.caption, design: .monospaced))
                                .foregroundStyle(Color.accentColor)

                            Spacer()

                            Button {
                                UIPasteboard.general.string = dialCode
                            } label: {
                                Image(systemName: "doc.on.doc")
                                    .font(.caption)
                                    .foregroundStyle(.secondary)
                            }

                            if let url = URL(string: "tel:\(dialCode)") {
                                Link(destination: url) {
                                    Text("Aktivovat")
                                        .font(.caption2)
                                        .padding(.horizontal, 10)
                                        .padding(.vertical, 6)
                                        .background(Color.accentColor)
                                        .foregroundStyle(.white)
                                        .clipShape(Capsule())
                                }
                            }
                        }
                    }

                    // Deactivation code
                    VStack(alignment: .leading, spacing: 4) {
                        Text("Kód pro zrušení")
                            .font(.caption)
                            .foregroundStyle(.secondary)

                        HStack {
                            Text(type.deactivateCode)
                                .font(.system(.caption, design: .monospaced))
                                .foregroundStyle(.red)

                            Spacer()

                            if let url = URL(string: "tel:\(type.deactivateCode)") {
                                Link(destination: url) {
                                    Text("Zrušit")
                                        .font(.caption2)
                                        .padding(.horizontal, 10)
                                        .padding(.vertical, 6)
                                        .background(Color.red.opacity(0.1))
                                        .foregroundStyle(.red)
                                        .clipShape(Capsule())
                                }
                            }
                        }
                    }
                }
                .padding([.horizontal, .bottom])
            }
        }
        .background(Color(.systemGray6))
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }
}

#Preview {
    RedirectQuickSetupView(karenNumber: "+420 123 456 789")
}
