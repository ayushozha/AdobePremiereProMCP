"""Literal media references must not be diluted or replaced by path overlap."""

from __future__ import annotations

import pytest

from src.matching.keyword_matcher import KeywordMatcher
from src.matching.matcher import AssetMatcher
from src.models import AssetInfo, AssetType, MatchStrategy, SegmentType
from src.parser.script_parser import ScriptParser

ROOT = (
    "/Users/editor/Desktop/01 Work & Ventures/Launch Week/Video/Video Editing/"
    "AdobePremiereProMCP-issue-workspaces/evidence/e2e-fixture/media"
)


def assets() -> list[AssetInfo]:
    return [
        AssetInfo(
            id="video", file_name="e2e_test_pattern.mp4",
            file_path=f"{ROOT}/e2e_test_pattern.mp4", asset_type=AssetType.VIDEO,
            metadata={"description": "unrelated broad metadata can never dilute a literal file"},
        ),
        AssetInfo(
            id="audio", file_name="e2e_tone.wav",
            file_path=f"{ROOT}/e2e_tone.wav", asset_type=AssetType.AUDIO,
        ),
        AssetInfo(
            id="similar", file_name="e2e_test_pattern_backup.mp4",
            file_path=f"{ROOT}/e2e_test_pattern_backup.mp4", asset_type=AssetType.VIDEO,
        ),
    ]


@pytest.mark.parametrize("strategy", [MatchStrategy.KEYWORD, MatchStrategy.HYBRID,
                                     MatchStrategy.EMBEDDING])
@pytest.mark.parametrize("reference", ["e2e_test_pattern.mp4", f"{ROOT}/e2e_test_pattern.mp4",
                                      "media/e2e_test_pattern.mp4"])
def test_literal_reference_matches_only_the_exact_asset_without_api_calls(
    strategy: MatchStrategy, reference: str, monkeypatch: pytest.MonkeyPatch,
) -> None:
    script = ScriptParser().parse(f'B-ROLL: "{reference}"\n', format_hint="youtube")
    assert [segment.type for segment in script.segments] == [SegmentType.BROLL]
    matcher = AssetMatcher(strategy=strategy)

    def unexpected_embedding(*args: object, **kwargs: object) -> None:
        pytest.fail("Literal references must not require an embedding API or fallback")

    monkeypatch.setattr(matcher.embedding_matcher, "match", unexpected_embedding)
    result = matcher.match(script.segments, assets())
    assert [(match.asset_id, match.confidence) for match in result.matches] == [("video", 1.0)]
    assert not result.unmatched
    # A shared parent folder or near-identical filename cannot substitute.
    missing = matcher.match(script.segments, assets()[1:])
    assert not missing.matches
    assert len(missing.unmatched) == 1


def test_absolute_path_disambiguates_identical_basenames() -> None:
    same_name = AssetInfo(id="other-folder", file_name="e2e_test_pattern.mp4",
                          file_path="/other-folder/e2e_test_pattern.mp4")
    segment = ScriptParser().parse(f'B-ROLL: "{ROOT}/e2e_test_pattern.mp4"').segments[0]
    assert [match.asset_id for match in KeywordMatcher().match(segment, [same_name, *assets()])] == [
        "video",
    ]


def test_nonexistent_absolute_path_does_not_fall_back_to_an_existing_basename() -> None:
    segment = ScriptParser().parse('B-ROLL: "/missing-folder/e2e_test_pattern.mp4"').segments[0]
    assert KeywordMatcher().match(segment, assets()) == []


def test_windows_paths_and_quoted_filenames_with_spaces_preserve_the_reference() -> None:
    asset = AssetInfo(id="named", file_name="My Test Pattern.MP4",
                      file_path=r"C:\Edit Project\media\My Test Pattern.MP4")
    for reference in [asset.file_name, asset.file_path, "media/My Test Pattern.MP4"]:
        segment = ScriptParser().parse(f'B-ROLL: "{reference}"').segments[0]
        assert [match.asset_id for match in KeywordMatcher().match(segment, [asset])] == ["named"]


def test_unquoted_filename_in_a_visual_description_is_an_explicit_reference() -> None:
    segment = ScriptParser().parse("B-ROLL: show e2e_test_pattern.mp4 for four seconds").segments[0]
    assert [match.asset_id for match in KeywordMatcher().match(segment, assets())] == ["video"]


def test_duplicate_basenames_fail_closed_instead_of_using_scan_order() -> None:
    duplicate = AssetInfo(id="other", file_name="e2e_test_pattern.mp4",
                         file_path="/other-folder/e2e_test_pattern.mp4")
    segment = ScriptParser().parse('B-ROLL: "e2e_test_pattern.mp4"').segments[0]
    for candidates in [[duplicate, *assets()], [*assets(), duplicate]]:
        assert KeywordMatcher().match(segment, candidates) == []


def test_actual_parser_preserves_posix_case_despite_lowercased_derived_hints() -> None:
    candidates = [AssetInfo(id="lower", file_name="a.mp4", file_path="/stock/a.mp4"),
                  AssetInfo(id="upper", file_name="A.mp4", file_path="/stock/A.mp4")]
    segment = ScriptParser().parse('B-ROLL: "/stock/A.mp4"').segments[0]
    assert "/stock/a.mp4" in segment.asset_hints  # Existing extractor is lossy.
    assert [match.asset_id for match in KeywordMatcher().match(segment, candidates)] == ["upper"]
    assert KeywordMatcher().match(segment, candidates[:1]) == []


def test_spaced_unquoted_path_cannot_bind_its_suffix_in_another_directory() -> None:
    candidates = [AssetInfo(id="wrong", file_name="clip.mp4",
                            file_path="/unrelated/folder/clip.mp4"),
                  AssetInfo(id="right", file_name="clip.mp4",
                            file_path="/wanted folder/clip.mp4")]
    segment = ScriptParser().parse("B-ROLL: /wanted folder/clip.mp4").segments[0]
    assert [match.asset_id for match in KeywordMatcher().match(segment, candidates)] == ["right"]
    assert KeywordMatcher().match(segment, candidates[:1]) == []


def test_quoted_filename_characters_are_not_stripped() -> None:
    candidates = [AssetInfo(id="wrong", file_name="clip).mp4", file_path="/stock/clip).mp4"),
                  AssetInfo(id="right", file_name="(clip).mp4", file_path="/stock/(clip).mp4")]
    segment = ScriptParser().parse('B-ROLL: "(clip).mp4"').segments[0]
    assert [match.asset_id for match in KeywordMatcher().match(segment, candidates)] == ["right"]


def test_multiple_literal_assets_require_separate_segments() -> None:
    segment = ScriptParser().parse('B-ROLL: "first.mp4" and "second.mp4"').segments[0]
    candidates = [AssetInfo(id="first", file_name="first.mp4", file_path="/stock/first.mp4"),
                  AssetInfo(id="second", file_name="second.mp4", file_path="/stock/second.mp4")]
    for ordered in [candidates, candidates[::-1]]:
        assert KeywordMatcher().match(segment, ordered) == []


def test_literal_scene_description_is_binding() -> None:
    segment = ScriptParser().parse('[SECTION: /missing/e2e_test_pattern.mp4]\nB-ROLL: footage').segments[0]
    assert KeywordMatcher().match(segment, assets()) == []


def test_contraction_apostrophes_cannot_hide_a_missing_literal_reference() -> None:
    description = "don't forget missing.mp4, it's required"
    segment = ScriptParser().parse("B-ROLL: " + description).segments[0]
    unrelated = AssetInfo(id="backup", file_name="backup.mp4",
                          metadata={"description": description})
    assert KeywordMatcher().match(segment, [unrelated]) == []


def test_one_available_reference_cannot_hide_a_second_missing_file() -> None:
    segment = ScriptParser().parse('B-ROLL: "e2e_test_pattern.mp4" and "missing.mp4"').segments[0]
    assert KeywordMatcher().match(segment, assets()) == []


@pytest.mark.parametrize("reference", ["wanted folder/clip.mp4", r"wanted folder\clip.mp4",
                                      "show wanted folder/clip.mp4",
                                      r"show wanted folder\clip.mp4"])
def test_ambiguous_unquoted_relative_path_requires_quotes(reference: str) -> None:
    unrelated = AssetInfo(id="wrong", file_name="clip.mp4",
                          file_path="/unrelated/folder/clip.mp4")
    segment = ScriptParser().parse("B-ROLL: " + reference).segments[0]
    assert KeywordMatcher().match(segment, [unrelated]) == []


def test_quoted_relative_path_with_spaces_remains_binding() -> None:
    candidates = [AssetInfo(id="wrong", file_name="clip.mp4",
                            file_path="/unrelated/folder/clip.mp4"),
                  AssetInfo(id="right", file_name="clip.mp4",
                            file_path="/stock/wanted folder/clip.mp4")]
    segment = ScriptParser().parse('B-ROLL: "wanted folder/clip.mp4"').segments[0]
    assert [match.asset_id for match in KeywordMatcher().match(segment, candidates)] == ["right"]


def test_generic_descriptions_still_use_existing_keyword_and_embedding_scoring(
    monkeypatch: pytest.MonkeyPatch,
) -> None:
    segment = ScriptParser().parse("B-ROLL: sunset beach").segments[0]
    candidate = AssetInfo(id="sunset", file_name="sunset_beach.mp4",
                          file_path="/stock/sunset_beach.mp4")
    matcher = AssetMatcher()
    matcher.embedding_matcher._available = False
    calls: list[bool] = []
    original = matcher.embedding_matcher.match

    def track_embedding(*args: object, **kwargs: object) -> object:
        calls.append(True)
        return original(*args, **kwargs)  # type: ignore[arg-type]

    monkeypatch.setattr(matcher.embedding_matcher, "match", track_embedding)
    assert KeywordMatcher.match_explicit_references(segment, [candidate]) is None
    result = matcher.match([segment], [candidate])
    assert calls == [True]
    assert result.matches[0].asset_id == "sunset"
    assert result.matches[0].confidence < 1.0
