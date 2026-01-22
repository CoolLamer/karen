import { describe, it, expect } from "vitest";
import {
  CLEAR_ALL_REDIRECTS_CODE,
  REDIRECT_CODES,
  REDIRECT_ORDER,
  NO_ANSWER_TIME_OPTIONS,
  DEFAULT_NO_ANSWER_TIME,
  getDialCode,
  getDescription,
  getDeactivationCode,
} from "./redirectCodes";

describe("redirectCodes constants", () => {
  describe("CLEAR_ALL_REDIRECTS_CODE", () => {
    it("should be the correct code to clear all redirects", () => {
      expect(CLEAR_ALL_REDIRECTS_CODE).toBe("##002#");
    });
  });

  describe("REDIRECT_CODES", () => {
    it("should have noAnswer type", () => {
      expect(REDIRECT_CODES.noAnswer).toBeDefined();
      expect(REDIRECT_CODES.noAnswer.code).toContain("61");
      expect(REDIRECT_CODES.noAnswer.deactivateCode).toBe("##61#");
    });

    it("should have busy type", () => {
      expect(REDIRECT_CODES.busy).toBeDefined();
      expect(REDIRECT_CODES.busy.code).toContain("67");
      expect(REDIRECT_CODES.busy.deactivateCode).toBe("##67#");
    });

    it("should have unreachable type", () => {
      expect(REDIRECT_CODES.unreachable).toBeDefined();
      expect(REDIRECT_CODES.unreachable.code).toContain("62");
      expect(REDIRECT_CODES.unreachable.deactivateCode).toBe("##62#");
    });
  });

  describe("REDIRECT_ORDER", () => {
    it("should have correct order", () => {
      expect(REDIRECT_ORDER).toEqual(["noAnswer", "busy", "unreachable"]);
    });
  });

  describe("NO_ANSWER_TIME_OPTIONS", () => {
    it("should have all time options", () => {
      expect(NO_ANSWER_TIME_OPTIONS).toEqual([5, 10, 15, 20, 25, 30]);
    });
  });

  describe("DEFAULT_NO_ANSWER_TIME", () => {
    it("should be 10 seconds", () => {
      expect(DEFAULT_NO_ANSWER_TIME).toBe(10);
    });
  });
});

describe("getDialCode", () => {
  const testPhone = "+420123456789";

  it("should generate noAnswer code with phone and time", () => {
    const code = getDialCode("noAnswer", testPhone, 15);
    expect(code).toBe("**61*+420123456789**15#");
  });

  it("should use default time for noAnswer when not provided", () => {
    const code = getDialCode("noAnswer", testPhone);
    expect(code).toBe("**61*+420123456789**10#");
  });

  it("should generate busy code with phone", () => {
    const code = getDialCode("busy", testPhone);
    expect(code).toBe("**67*+420123456789#");
  });

  it("should generate unreachable code with phone", () => {
    const code = getDialCode("unreachable", testPhone);
    expect(code).toBe("**62*+420123456789#");
  });

  it("should strip spaces from phone number", () => {
    const code = getDialCode("busy", "+420 123 456 789");
    expect(code).toBe("**67*+420123456789#");
  });
});

describe("getDescription", () => {
  it("should return noAnswer description with time placeholder replaced", () => {
    const desc = getDescription("noAnswer", 20);
    expect(desc).toContain("20 sekund");
  });

  it("should use default time when not provided for noAnswer", () => {
    const desc = getDescription("noAnswer");
    expect(desc).toContain("10 sekund");
  });

  it("should return busy description", () => {
    const desc = getDescription("busy");
    expect(desc).toContain("obsazeno");
  });

  it("should return unreachable description", () => {
    const desc = getDescription("unreachable");
    expect(desc).toContain("signál");
  });
});

describe("getDeactivationCode", () => {
  it("should return correct code for noAnswer", () => {
    expect(getDeactivationCode("noAnswer")).toBe("##61#");
  });

  it("should return correct code for busy", () => {
    expect(getDeactivationCode("busy")).toBe("##67#");
  });

  it("should return correct code for unreachable", () => {
    expect(getDeactivationCode("unreachable")).toBe("##62#");
  });
});
